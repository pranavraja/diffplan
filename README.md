
Generate a structured array of changes for automation (CREATE, UPDATE, DELETE) just by editing raw data as [logfmt](https://brandur.org/logfmt)-style records in a text file.

Note: This is still a work in progress, and is only used by one dev (me) at the moment, so although I've tried to make very few assumptions here, the features are probably quite specific to my workflow.

## Background

I often write custom scripts to make automated changes to sets of remote resources like cloud infrastructure, jira ticket statuses, database records, etc.

There's a common pattern with so many of these that I end up writing again and again, like writing custom imperative logic to check specific fields on certain resources, printing proposed changes out in a "dry run" mode, realising my if statements were slightly off, tweaking the script a few times until the dry run is what I expect, then passing in a flag to execute the changes, etc.

I wish I could just:

- see a list of resources I'm working with, their current status and other fields
- just use my text editor to filter and edit them into the state I want
- save, and see a preview of all the changes I made
- pass it to a very simple "execute" script I prepared earlier to make the changes

## Setup

You'll need [fzf](https://github.com/junegunn/fzf) installed to take full advantage of the command completion.

Set your `EDITOR` to your preferred editor (defaults to `vim`, but you can set `subl -w`, for example)

To add this binary to your `$GOPATH/bin`, just run:

	go install github.com/pranavraja/diffplan@master

## How to use this

- Prepare a text file `input.txt` with a bunch of lines, where each line represents one of the resources you want to manage, and an arbitrary set of fields for each, e.g.:

```
id=i-abcd1234 name=worker1 state=running
id=i-lkjh4321 name=worker2 state=running
id=i-poiu5678 name=testinstancecostingmemoney state=running
id=i-qwer8765 name=anothertestiforgotabout state=running
```

(you'd probably want the above data file to be the output of something automated, like a program that lists your ec2 instances and prints a line for each one)

- Run `diffplan input.txt`. This will bring up your editor and let you modify the resources.

- Make whatever changes you want (e.g. maybe in vim you do `:%!grep -v test` to remove all the test stuff, and find `worker2` and add a new field `tag=autoscaling`)

- Save and close the editor, and diffplan will print out an color-coded list of 3 changes, an `UPDATE` to worker2, and a `DELETE` for your 2 test instances:

```
will UPDATE:
 - #i-lkjh4321: tag=autoscaling

will DELETE:
 - #i-poiu5678: name=testinstancecostingmemoney state=running
 - #i-qwer8765: name=anothertestiforgotabout state=running
```

- If you continue, you can then type in a separate "executor" command to actually make these changes. The executor command will receive the following JSON on stdin:

```json
[{"change":"UPDATE","id":"i-lkjh4321","fields":{"tag":"autoscaling"},"old":{"name":"worker2","state":"running"}},{"change":"DELETE","id":"i-poiu5678","fields":{"name":"testinstancecostingmemoney","state":"running"}},{"change":"DELETE","id":"i-qwer8765","fields":{"name":"anothertestiforgotabout","state":"running"}}]
```

You can use your own rules to decide what the executor command does with the list of changes. Maybe for DELETEs you just want to *stop* any deleted instances instead of terminating them. Maybe for UPDATES, depending on the fields passed in, you want to call a different API. It's up to you.

As a bonus feature, a list of previous executor commands you've typed when in the current directory is maintained in a file `.plan-execute-history` so you can re-run common executor commands you've used.

## Examples to pique your interest

Think about how you would achieve the below with your existing setup.

1. I want to see all the ec2 instances in my account that are untagged, go through them and tag them appropriately based on their name/keyname/whatever if I recognise them, and stop or delete them if their names contain `-test-`, or another criteria I think of when I see the actual data.

2. I want to see all the jira tickets I'm currently watching, stop watching the ones I don't care about, and assign all the ones I do care about to myself.

## FAQ

### Why would I use this? I have to write adapter code anyway for input and output, so I might as well just write the whole thing myself

Yeah fair point. Maybe you find the idea of planning and previewing changes as a separate step useful, even if you don't need the actual code, which is only a few hundred lines.

### Isn't this what declarative configuration tools like terraform/ansible/whatever are for?

Kinda but not really. The extra setup of naming and declaring all your resources and desired statuses upfront is a bit overkill for me. Plus most other tools make assumptions about how to apply your changes, i.e. they have to know what an UPDATE/DELETE actually does. This tool operates purely on data, and doesn't care what you do with the list of changes.

### Can I rely on the structured output format to be constant? So I can build a library of executor scripts using this format?

Ehhh probably. I might add more fields if I need to but the format should stay the same.

