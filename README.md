# `credhub-service-broker`

This is a Cloud Foundry service broker allowing users to store app
secrets in CredHub. That offers more effective security and auditing
than storing secrets in Cloud Controller.

Secrets stored in app environment variables (`cf set-env`) or in
user-provided services (`cf cups`) can be seen in `cf env` or
`/v2/app/:guid` responses. Auditing of those endpoints is ineffective
(they have many regular legitimate uses.)

When using this service broker, secrets should only be accessible by:

  - SSHing into an app, generating CC Audit Events

  - accessing CredHub, generating CredHub Audit Events

  - having the app print them out, which can be mitigated by code
    review

  - compromising the platform or the app

## Usage

Here's an example of how to use it to provide AWS credentials to apps:

```
$ cf create-service secrets simple aws-secrets \
    -c '{"AWS_ACCESS_KEY_ID": "…", "AWS_SECRET_ACCESS_KEY": "…", […]}'

$ cf push first-app-using-aws-creds --no-start
$ cf bind-service first-app-using-aws-creds aws-secrets
$ cf start first-app-using-aws-creds

$ cf push second-app-using-aws-creds --no-start
$ cf bind-service second-app-using-aws-creds aws-secrets
$ cf start second-app-using-aws-creds
```

If you run `cf env` against those apps you don't get back anything
sensitive:

```
VCAP_SERVICES='{
  "secrets": [{
    "credentials": {
      "credhub-ref": "non-sensitive-string"
    }
  }]
}'
```

The value of `credhub-ref` is only useful if you already have access
to CredHub, and CredHub access can be audited well.

Cloud Foundry retrieves the credentials from CredHub when the apps
start, without any changes needed to your app. The app sees the proper
AWS credentials in `VCAP_SERVICES`:

```
VCAP_SERVICES='{
  "secrets": [{
    "credentials": {
      "AWS_ACCESS_KEY_ID": "AZ…"
      "AWS_SECRET_ACCESS_KEY": "Zy…",
      […]
    }
  }]
}'
```

As usual for CF apps your app can either parse the `VCAP_SERVICES`
environment variable or use something like the
[env-map-buildpack](https://github.com/andy-paine/env-map-buildpack)
to do it for you at app startup.

## Installation

First your Cloud Foundry deployment has to be able to support this:

* You need to have enabled CredHub in your Cloud Foundry deployment
* You need to have configured a `credhub` link to Cloud Controller in
  your BOSH manifest
* You need to allow traffic from all apps to the VMs hosting CredHub,
  for instance by adding the appropriate CF Security Groups
* You need to allow the `credhub-service-broker` app to talk to your
  CredHub, UAA and Locket
* There is an example of how we did some of this for our Cloud Foundry
  deployment: [commit 5761b8df on github.com/alphagov/paas-cf](https://github.com/alphagov/paas-cf/commit/5761b8df854bff3d768c3d68b39b075180c14840)

Once all that is done you need to collect quite a few configuration
variables. Make a copy of `blank-vars-store.yml` and fill it out with
values for your Cloud Foundry deployment. Then try deploying it:

```
$ cf push --vars-file vars-store.yml

$ cf create-service-broker \
    credhub-broker \
    BASIC-AUTH-USERNAME-FROM-VARS-FILE \
    BASIC-AUTH-PASSWORD-FROM-VARS-FILE \
    ROUTE-FROM-VARS-FILE

$ cf enable-service-access -b credhub-broker
```

If the configuration works you should now be able to follow the usage
examples further up this document.

So far there are no acceptance or smoke tests. Those will come later
if this project gets adopted.

## Future direction

So far this broker re-uses existing Cloud Foundry features created for
Pivotal's closed-source CredHub Service Broker. However that comes
with some big limitations that prevent our users from taking full
advantage of CredHub's automation (e.g., acting as a CA.) In future
we could grant our users directly access to Credhub so they can [manage
secrets with Terraform](https://github.com/orange-cloudfoundry/terraform-provider-credhub)
or similar. We could also support the
[credhub-buildpack](https://github.com/andy-paine/credhub-buildpack)
to allow more flexible use of secrets.
