package fetch

import (
	"context"
	"fmt"
	"strings"
	"time"

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

func AliasToAccountMap(ctx context.Context, masterProfile, accountRole string) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		Profile:                 masterProfile,
	}))

	oal, err := GetAWSOrganizationsAccounts(ctx, sess)
	if err != nil {
		common.ExitWithError(err)
	}

	al := GetAccounts(oal)

	if err = GetTagsForOU(ctx, sess, al); err != nil {
		common.ExitWithError(err)
	}

	if err = GetAliases(ctx, sess, al, accountRole); err != nil {
		common.ExitWithError(err)
	}

	common.WriteAccountList(al)
}

func GetAWSOrganizationsAccounts(ctx context.Context, sess client.ConfigProvider) (oal []*organizations.Account, err error) {
	svc := organizations.New(sess)

	var o *organizations.ListAccountsOutput
	var nextToken *string
	for {
		if err = common.CheckContext(ctx); err != nil {
			return nil, err
		}

		o, err = svc.ListAccounts(&organizations.ListAccountsInput{
			MaxResults: MaxResults,
			NextToken:  nextToken,
		})
		if err != nil {
			_, cancel := context.WithCancel(ctx)
			cancel()
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

func GetAliases(ctx context.Context, sess client.ConfigProvider, al []*common.Account, accountRole string) (err error) {
	eg, ctx := errgroup.WithContext(ctx)

	// Send a maximum of 10 concurrent requests to AWS at a time.
	s := make(chan int, 10) // makeshift semaphore
	for i, a := range al {
		loopA := a
		s <- i
		eg.Go(func() error {
			time.Sleep(time.Second) // Slow things down to avoid rate limits.
			e := GetAlias(sess, loopA, accountRole)
			<-s
			return e
		})
		fmt.Printf("\rFetched aliases for %d accounts...", i+1)

		if err = common.CheckContext(ctx); err != nil {
			return
		}
	}

	fmt.Println()

	if err = eg.Wait(); err != nil {
		return
	}
	return
}

func GetAlias(sess client.ConfigProvider, a *common.Account, accountRole string) (err error) {
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

func GetTagsForOU(ctx context.Context, sess client.ConfigProvider, al []*common.Account) (err error) {
	eg, ctx := errgroup.WithContext(ctx)

	// Send a maximum of 1 concurrent requests to AWS at a time. Perhaps one
	// day, this loop can be used to send more than one request at a time.
	s := make(chan int, 1) // makeshift semaphore
	for i, a := range al {
		loopA := a
		s <- i
		eg.Go(func() error {
			e := GetTagsForAccount(ctx, sess, loopA)
			<-s
			return e
		})
		fmt.Printf("\rFetched tags for %d accounts...", i+1)

		if err = common.CheckContext(ctx); err != nil {
			return
		}
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
		if err = common.CheckContext(ctx); err != nil {
			return
		}

		o, err = svc.ListTagsForResource(&organizations.ListTagsForResourceInput{
			ResourceId: &a.Id,
			NextToken:  nextToken,
		})
		if err != nil {
			_, cancel := context.WithCancel(ctx)
			cancel()
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
