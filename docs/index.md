# About

!!! info "About `scm-engine`"

    SCM Engine allow for easy Merge Request automation within your GitLab projects.

    Automatically [add/remove labels](configuration/index.md#label) depending on files changes, the age of the Merge Request, who contributes, and pretty much anything else you could want.

    You can even [*take actions*](configuration/index.md#actions) such as ([but not limited to](configuration/index.md#actions.if.then.action)) closing the Merge Request, approve it, or add a comment.

    SCM engine can be run either as a [regular CI job in your pipeline](gitlab/setup.md#gitlab-ci-pipeline), or be [triggered through the Webhook system](gitlab/setup.md#webhook-server), allowing for versatile and flexible deployments.
