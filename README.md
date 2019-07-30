# Slack Status

A command line app that allows you to quickly set your Slack status according to predefined templates.
For instance, you can mark your status as out to lunch or working from home with `slackstatus lunch`
or `slackstatus wfh` respectively.

## Features

- Customize which emoji is used for your status and the status text to your needs.
- Add your own status templates such as being at a doctor's appointment or on maternity leave.
- Supports multiple workspaces to either set your status at all of your workspaces or just one.
- You can download the binary or compile it yourself (if you have Go installed) with `./build.sh`.

## How to Use

Put the `slackstatus` app from `bin/slackstatus` anywhere that is within your `$PATH`.

The format for commands is `slackstatus [workspace] command` where `workspace` is an optional short
name for the Slack workspace you want to change the status of. Omitting the workspace parameter will
apply your status to _all_ of your workspaces in the config file.

`command` is your short name for your
status, some of the defaults being:

- `work`
- `done`
- `wfh`
- `lunch`

## To Do

- [ ] Investigate integrating Cobra for better help output
- [ ] Better output aside from raw request data
- [ ] Multi-thread requests
- [x] Support configs
