# terraform-provider-dagster
Terraform provider to manage Dagster Cloud resources.

> **Warning**
> This is a very early version: it implements just what we needed to manage our Dagster organization with Terraform.

## Coverage

| Type | Implemented as Resource | Implemented as Data Source |
| ------------- | ------------- | ------------- |
| User | :heavy_check_mark: | :heavy_check_mark: |
| Team  |  :heavy_check_mark: | :heavy_check_mark: |
| Team permission on Deployment  |  :heavy_check_mark: | |
| Current deployment | | :heavy_check_mark: |
| Code location | Partial | :x: |
| Team | :heavy_check_mark: | :x: |
| Team membership | :heavy_check_mark: | :x: |


## Roadmap

- [ ] Deployment / Deployment Settings
- [ ] Release to terraform provider registry
- [ ] Configure provider with env vars

## Useful links
- gql code gen: https://github.com/Khan/genqlient & https://github.com/Khan/genqlient/blob/main/docs/introduction.md
- dagster gql api: https://\<instance\>.dagster.cloud/\<deployment\>/graphql
- dagster python cloud sdk: https://github.com/dagster-io/dagster-cloud
