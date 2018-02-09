#!/bin/bash


composer runtime install --card PeerAdmin@hlfv1 --businessNetworkName chatroom

composer network start --card PeerAdmin@hlfv1 --networkAdmin admin --networkAdminEnrollSecret adminpw --archiveFile chatroom@0.0.1.bna --file networkadmin.card
