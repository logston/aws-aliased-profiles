# aws-aliased-profiles

Issue:

- You have a bunch of accounts in an organization and you want 

```
aws-aliased-profiles fetch -masterProfile root -accountRole Production
```


Build map of Id's to aliases

This part ^^^ should be a pkg so its importable


This part VVVV should be a CLI. Two repos???

use that to create alias named profiles

it will need to take the profile to fetch the accounts with, default may not always be the right profile

Insert named profiles into `~/.aws/config` file
  - in between # ------ aws-named-profiles ------ start and end tags
  - Alert if file does not exist
  - Should handle config being in other dirs than ~/.aws

Command should be able to redownload alias listing
Command should be able to regenerate map

profile_template = ```
[profile {{ alias }}]
role_arn = arn:aws:iam::{{ accountId }}:role/MyFavRoleToAssume
source_profile = default
```

Config files
	- ~/.aws/named-profiles/config
		Holds template to go under `[profile]` section for each account

Product should have good tests
