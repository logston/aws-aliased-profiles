package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/organizations"
	"golang.org/x/sync/errgroup"
)

// MaxResults defined by API is 20
var MaxResults = aws.Int64(int64(20))

type Account struct {
	// The unique identifier (ID) of the account.
	//
	// The regex pattern (http://wikipedia.org/wiki/regex) for an account ID string
	// requires exactly 12 digits.
	Id string

	// The date the account became a part of the organization.
	JoinedTimestamp time.Time

	// The status of the account in the organization.
	Status string

	// Alias associated with the account.
	Alias string
}

func main() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
	}))

	oal, err := GetOrganizationAccounts(sess)
	if err != nil {
		Exit(err)
	}

	al := GetAccounts(oal)

	profile := "MyRole"
	if err = PopulateAliases(sess, al, profile); err != nil {
		Exit(err)
	}

	PrintAccounts(al)
}

func Exit(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func GetOrganizationAccounts(sess client.ConfigProvider) (oal []*organizations.Account, err error) {
	svc := organizations.New(sess)

	var o *organizations.ListAccountsOutput
	var nextToken *string
	i := int64(1)
	for {
		o, err = svc.ListAccounts(&organizations.ListAccountsInput{
			MaxResults: MaxResults,
			NextToken:  nextToken,
		})
		if err != nil {
			return
		}
		fmt.Printf("\rFetched %d...", i*(*MaxResults))

		oal = append(oal, o.Accounts...)

		if o.NextToken == nil {
			break
		}

		nextToken = o.NextToken
		i += 1

		if i > 5 {
			break
		}
	}
	fmt.Println()

	return
}

func GetAccounts(oal []*organizations.Account) (al []*Account) {
	for _, oa := range oal {
		al = append(al, &Account{
			Id:              *oa.Id,
			JoinedTimestamp: *oa.JoinedTimestamp,
			Status:          *oa.Status,
		})
	}
	return
}

func PopulateAliases(sess client.ConfigProvider, al []*Account, profile string) (err error) {
	eg, ctx := errgroup.WithContext(NewCtx())

	for _, a := range al {
		loopA := a
		eg.Go(func() error {
			return PopulateAlias(ctx, sess, loopA, profile)
		})
	}

	if err = eg.Wait(); err != nil {
		return
	}
	return
}

func PopulateAlias(ctx context.Context, sess client.ConfigProvider, a *Account, profile string) (err error) {
	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", a.Id, profile)
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

func PrintAccounts(al []*Account) {
	for _, a := range al {
		fmt.Printf("%s -> %s, %s %s\n", a.Id, a.Alias, a.JoinedTimestamp, a.Status)
	}
}
