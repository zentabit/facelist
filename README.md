Facelist
========

A GAE app that queries Slack API for user profiles and show them on a simple web page.

Configuration
=============

facelist is configured by setting environment variables in app.yaml:

    * EMAIL_FILTER - Only users with a email ending with this string will be showed
    * SLACK_TEAM - The name of your slack team
    * SLACK_API_TOKEN - Access token to the Slack api

Development
===========

Download external dependencies

    $ go mod download

Install google-cloud-sdk

    $ brew cask install google-cloud-sdk

Run locally:

    $ dev_appserver.py app.yaml

The facelist should be served at http://localhost:8080/

Deploy app
==========

    $ gcloud config set project <gae project id>
    $ gcloud app deploy