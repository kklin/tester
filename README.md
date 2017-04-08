# Quilt CI

This repository contains a Quilt spec for a Jenkins deployment that tests
[Quilt](github.com/quilt/quilt).

It also contains a custom Docker image because we require some build tools and
Jenkins plugins for our testing job.

Note that by default the quilt-tester job has the `--preserve-failed` flag enabled.
This flag leaves machines up if there were any test failures to facilitate debugging.
Thus, if running locally, you may have to manually destroy the quilt-tester machines
if your final test run ends in a failure.

## Deploying
`tester.js` contains a module that creates a Jenkins service. It requires some
configuration options to be passed in by the caller. An example is included at
[tester-runner-example.js](tester-runner-example.js).

The required configurations are:
- `awsAccessKey`: The access key used by test daemon to boot machines.
- `awsSecretAccessKey`: The secret access key used by test daemon to boot machines.
- `testingNamespace`: The namespace used by the test daemon.
- `slackChannel`: The Slack channel where build results will be posted.
- `slackTeam`: The Slack team in which `slackChannel` is located.
- `slackToken`: The token used for posting to `slackChannel`.

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
