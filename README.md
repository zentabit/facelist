Facelist
========

An app that queries Slack API for user profiles and show them on a simple web page.

Configuration
=============

facelist is configured by a config file e.g facelist.yaml:

    * emailFilter - Only users with a email ending with this string will be showed
    * slackTeam - The name of your slack team
    * slackAPIToken - Access token to the Slack api

The API-token requires the scopes:
* users:read
* users:read.email

Development
===========

Download external dependencies

    $ go mod download

Build and run locally:

    $ go build
    $ ./facelist

The facelist should be served at http://localhost:8080/

Deploy app
==========
The included dockerfile can be used to deploy the app.
