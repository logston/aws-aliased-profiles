package fetch

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/organizations"
	"golang.org/x/sync/errgroup"

	"github.com/logston/aws-aliased-profiles/common"
)

// MaxResults defined by API is 20
var MaxResults = aws.Int64(int64(20))

func AliasToAccountMap(masterProfile, accountRole string) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		Profile:                 masterProfile,
	}))

	oal, err := GetAWSOrganizationsAccounts(sess)
	if err != nil {
		common.ExitWithError(err)
	}

	al := GetAccounts(oal)

	if err = GetTagsForOU(sess, al); err != nil {
		common.ExitWithError(err)
	}

	if err = GetAliases(sess, al, accountRole); err != nil {
		common.ExitWithError(err)
	}

	common.WriteAccountList(al)
}

func GetAWSOrganizationsAccounts(sess client.ConfigProvider) (oal []*organizations.Account, err error) {
	svc := organizations.New(sess)

	var o *organizations.ListAccountsOutput
	var nextToken *string
	for {
		o, err = svc.ListAccounts(&organizations.ListAccountsInput{
			MaxResults: MaxResults,
			NextToken:  nextToken,
		})
		if err != nil {
			return
		}

		oal = append(oal, o.Accounts...)
		fmt.Printf("\rFetched %d accounts...", len(oal))

		if o.NextToken == nil {
			break
		}

		nextToken = o.NextToken
	}
	fmt.Println()

	return
}

func GetAccounts(oal []*organizations.Account) (al []*common.Account) {
	for _, oa := range oal {
		al = append(al, &common.Account{
			Id:              *oa.Id,
			JoinedTimestamp: *oa.JoinedTimestamp,
			Status:          *oa.Status,
		})
	}
	return
}

func NewCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sCh := make(chan os.Signal, 1)
		signal.Notify(sCh, syscall.SIGINT, syscall.SIGTERM)
		<-sCh
		cancel()
	}()
	return ctx
}

func GetAliases(sess client.ConfigProvider, al []*common.Account, accountRole string) (err error) {
	eg, ctx := errgroup.WithContext(NewCtx())

	// Send a maximum of 20 concurrent requests to AWS at a time
	s := make(chan int, 20)
	for i, a := range al {
		loopA := a
		s <- i
		eg.Go(func() error {
			e := GetAlias(ctx, sess, loopA, accountRole)
			<-s
			return e
		})
		fmt.Printf("\rFetched aliases for %d accounts...", i+1)
	}

	fmt.Println()

	if err = eg.Wait(); err != nil {
		return
	}
	return
}

func GetAlias(ctx context.Context, sess client.ConfigProvider, a *common.Account, accountRole string) (err error) {
	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", a.Id, accountRole)
	creds := stscreds.NewCredentials(sess, roleArn)
	svc := iam.New(sess, &aws.Config{Credentials: creds})

	var o *iam.ListAccountAliasesOutput
	o, err = svc.ListAccountAliases(&iam.ListAccountAliasesInput{})
	if err != nil {
		if strings.HasPrefix(err.Error(), "AccessDenied") {
			return nil
		}
		return
	}

	if len(o.AccountAliases) == 1 {
		a.Alias = *o.AccountAliases[0]
	}

	return
}

func GetTagsForOU(sess client.ConfigProvider, al []*common.Account) (err error) {
	eg, ctx := errgroup.WithContext(NewCtx())

	// Send a maximum of 1 concurrent requests to AWS at a time. Perhaps one
	// day, this loop can be used to send more than one request at a time.
	s := make(chan int, 1)
	for i, a := range al {
		loopA := a
		s <- i
		eg.Go(func() error {
			e := GetTagsForAccount(ctx, sess, loopA)
			<-s
			return e
		})
		fmt.Printf("\rFetched tags for %d accounts...", i+1)
	}

	fmt.Println()

	if err = eg.Wait(); err != nil {
		return
	}
	return
}

func GetTagsForAccount(ctx context.Context, sess client.ConfigProvider, a *common.Account) (err error) {
	svc := organizations.New(sess)

	var o *organizations.ListTagsForResourceOutput
	var nextToken *string
	var tags []*common.Tag
	for {
		o, err = svc.ListTagsForResource(&organizations.ListTagsForResourceInput{
			ResourceId: &a.Id,
			NextToken:  nextToken,
		})
		if err != nil {
			return
		}

		for _, t := range o.Tags {
			newTag := &common.Tag{
				Key:   *t.Key,
				Value: *t.Value,
			}
			tags = append(tags, newTag)
		}

		if o.NextToken == nil {
			break
		}

		nextToken = o.NextToken
	}

	a.Tags = tags

	return
}
