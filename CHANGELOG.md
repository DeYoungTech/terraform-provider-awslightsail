# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2022-01-12

Initial Release! This release contains all of the resources that existed in the `aws` provider along with the addition of the `awslightsail_database` resource.

The AWSLightsail provider has been written using `aws-sdk-go-v2` and follows the new documentation standards for the terraform provider registry.

### Added Resources

* `database`
* `domain`
* `instance`
* `key_pair`
* `static_ip`
* `static_ip_attachment`

[0.1.0]: https://github.com/DeYoungTech/terraform-provider-awslightsail/releases/tag/v0.1.0