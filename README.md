# Kelda CI

This repository contains a Kelda blueprint for a Jenkins deployment that tests
[Kelda](github.com/kelda/kelda).

It also contains a custom Docker image because we require some build tools and
Jenkins plugins for our testing job.

Note that by default we enable the `--preserve-failed` for the integration-tester.
This flag leaves machines up if there were any test failures to facilitate debugging.
Thus, if running locally, you may have to manually destroy the integration-tester
machines if your final test run ends in a failure.

## Deploying
`tester.js` contains a module that creates a Jenkins service. It requires some
configuration options to be passed in by the caller. An example is included at
[tester-runner-example.js](tester-runner-example.js).

The required configurations are:
- `awsAccessKey`: The access key used by the test daemon to boot machines on
the Amazon provider.
- `awsSecretAccessKey`: The secret access key used by the test daemon to boot
machines on the Amazon provider.
- `awsS3AccessKey`: The access key to pass as a Kelda secret to tests that
read data from S3. This key needs permissions to read and list the bucket used
by the Spark integration test.
- `awsS3SecretAccessKey`: The secret access key corresponding to
`awsS3AccessKey`.
- `digitalOceanKey`: The secret key used to boot machines on the DigitalOcean
provider.
- `gceProjectID`: The ID of the project in which Google machines will be booted.
- `gcePrivateKey`: The private key of the service account used to boot machines
on the Google provider.
- `gceClientEmail`: The email address of the service account used to boot
machines on the Google provider.
- `testingNamespacePrefix`: The prefix for the namespace used by the test daemon.
- `slackChannel`: The Slack channel where build results will be posted.
- `slackWebhook`: The Slack webhook used to post the build results. This can be
- generated in the Slack workspace settings. For example, the Kelda Slack
webhook was generated here https://kelda.slack.com/apps/A0F7XDUAZ-incoming-webhooks.

The optional argments are:
- `passwordHash`: A Jenkins-compatible hash of the desired password for the `admin` user.
    - If this is not set, the `admin` user will have to be manually setup when
    accessing the Jenkins instance for the first time.
- `jenkinsUrl`: The URL at which this Jenkins deployment will be able to be accessed.
    - If this is not set, the Jenkins URL will have to be [manually
      setup](#manually-setting-the-jenkins-url) in order for Slack notification
      links to work.
    - This is usually a floating IP assigned to the Jenkins machine.

### Configuring Slack
The [Jenkins CI App](https://slack.com/apps/A0F7VRFKN-jenkins-ci) needs to be enabled
in the Slack team to which build notifications will be posted.

The `slackToken` is generated by installing the app.

### Manually Setting the Jenkins URL
If `jenkinsUrl` is not set, we need to manually configure it for Slack
notifications to generate clickable links.

To do so, navigate to `Jenkins -> Manage Jenkins -> Configure System`, and simply
hit `Save`.
