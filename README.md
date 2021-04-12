# aws-aliased-profiles

### The Issue:

You have a bunch of accounts in an organization. You want to be able to use the
--profile flag easily and don't want to add profiles to the ~/.aws/config file
for each. This tool fetches the accounts in your organization, fetches the
aliases associated with each account in your organization, and then inserts
the profiles necessary into your config file.

### Installation


### Usage

1. Initialize the `~/.aws/aliased-profiles/config.tmpl` file.

```sh
aws-aliased-profiles init
```

This places the default profile template contents into the file at `~/.aws/aliased-profiles/config.tmpl`.

1. To fetch all accounts in your organization and their aliases, run the following command:

```sh
# aws-aliased-profiles fetch <aws profile with organization access> <role to assume>

aws-aliased-profiles fetch default Production
```

The `<aws profile with organization access>` argument specifies the profile in
your `~/.aws/config` file to use for access AWS API calls.

The `<role to assume>` argument specifies the role to assume when getting STS
tokens for alias retrieval in each child account. For example, something like
ReadOnly, Production, ProductionAdmin, etc. Each team names this according to
their own style.

1. The upsert command uses the downloaded account IDs and aliases to build new
profiles and insert them into the ~/.aws/config file.

```sh
aws-aliased-profiles upsert
```

The profiles inserted into the `~/.aws/config` file are generated by populating
a template file at `~/.aws/aliased-profiles/config.tmpl`. You need to place
something like the following in the file named above. You will need to change
MyFavRoleToAssume to the role you want to assume when using the profile. Often,
this is the same profile used in the `fetch` command.

```
[profile {{ alias }}]
role_arn = arn:aws:iam::{{ accountId }}:role/MyFavRoleToAssume
source_profile = default
```

After that, you should be able to use all your profiles readily...

```sh
aws --profile staging-123 sts get-caller-identity
{
    "UserId": "ABCDEFGHIJKLMNOP:botocore-session-1234567890",
    "Account": "987654321",
    "Arn": "arn:aws:sts::987654321234:assumed-role/MyFavRoleToAssume/botocore-session-1234567890"
}
```

###### TODO
- Tests
- Make sections importable by other packages
