# SQS Consumer plugin

The `sqs_consumer` input plugin reads messages from AWS SQS queue and creates metrics
using one of the supported [input data formats](/docs/DATA_FORMATS_INPUT.md).

### Configuration

```
# Read metrics from SQS
[[inputs.sqs]]
  # add your sqs you want to subscribe here
  url = "https://sqs.eu-west-1.amazonaws.com/000000000000/telegraf-test"
  access_key = "ACCESS_KEY"
  secret_key = "SECRET_KEY"
  region = "eu-west-1"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "json"
```


### AWS Credentials

This plugin uses a credential chain for Authentication with the Sqs
API endpoint. In the following order the plugin will attempt to authenticate.
1. Assumed credentials via STS if `role_arn` attribute is specified (source credentials are evaluated from subsequent rules)
2. Explicit credentials from `access_key`, `secret_key`, and `token` attributes
3. Shared profile from `profile` attribute
4. [Environment Variables](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#environment-variables)
5. [Shared Credentials](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#shared-credentials-file)
6. [EC2 Instance Profile](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html)


recommended: you should create an IAM User which is part of a policy limited to that resource.
Example policy:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "sqs_consumer:DeleteMessage",
                "sqs_consumer:ReceiveMessage"
            ],
            "Effect": "Allow",
            "Resource": "arn:aws:sqs_consumer:eu-west-1:000000000000:telegraf-test"
        }
    ]
}
```

Permissions:
- ReceiveMessage 
    - required for reading messages from the queue
- DeleteMessage
    - required for batch deletion of processed messages from queue
    
