# Kind filters

Kind can be either just the kind or prefixed with the api version separated with slash.

Examples:

    --kind Deployment
    --kind networking.k8s.io/v1/Ingress

As a special case if the option end with / all the resourcces with that api version are considered:

    --kind apps/v/

To choose several kinds you just repat the option. For example

    -k Deployment -k CronJob