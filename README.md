# Littr.me

This project represents a new attempt at the social link aggregator paradigm.

It leverages the [ActivityPub](https://www.w3.org/TR/activitypub) web specification for social networking
and uses federation to prevent some of the problems that exist currently in similar communities.

<!-- The main problem it tries to solve is the dissolution over time of the community's interests and its split
into groups with tighter focused interests and dissenting  -->

The main reason why link aggregators lose their appeal is the increasing number of users, 
which more often than not, do not share the existing group's interests and philosophies. 
As the number grows, the topics start to expand over a larger spectrum, people get flooded with irrelevant content, 
while what is relevant gets buried. 
<!--From the an old member's perspective it's the "eternal September" effect.--> 

This is why the main component which will be missing in this implementation is the concept of using 
the same instance for merging multiple interest groups and replacing it with the different instances themselves.
This way every community will be able to create one of their own  and enforce (or not) 
their own moderation rules and topic preferences.

At the same time, through the federation mechanism between instances, the users can subscribe to other
streams and will receive updates from them. As a plus, due to the interoperability that ActivityPub brings,
they are not only limited to link aggregators, but can also interact with other AP capable services.

# Installing

## Pre-requisites

The basic requirements for running littr.go locally are a postgresql server (version 9.5 or newer)
and a go environment (version 1.11 or newer). 

    $ go get github.com/mariusor/littr.go
    $ cp .env.example .env
    $ $EDITOR .env

In the `.env` file you can set-up the connection to the postgresql database and various other knobs and switches
to configure the application.

## Running 

    $ make run
