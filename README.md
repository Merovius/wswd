# wswd - workspace workdir

`wswd` is a tiny utility for [i3](https://i3wm.org/), which allows you to set a
working directory for a workspace. It is similar in nature to
[wsmgr](https://github.com/stapelberg/wsmgr-for-i3), but for a more dynamic
workflow.

The association between workspaces and directories is ephemeral. That is, it
does not survive i3 restarts and it goes away if the workspace gets destroyed
(e.g. because you do not have it assigned and it has no containers). The exact
details of when that happens are unspecified.

# Installation

You need a recent Go installation. Install `wswd` via `go install
gonih.org/wswd@latest`.  Also, make sure `$GOPATH/bin` or `$GOBIN` is in your
`$PATH`.

# Usage

`wswd` has four subcommands:

- `wswd set`: Stores the current working directory for the current workspace.
- `wswd unset`: Clears the working directory for the current workspace.
- `wswd exec <cmd…>`: Run `<cmd>` in the working directory associated with the
  current workspace.
- `wswd clean`: Removes directory associations for non-existent workspaces.

To set it up, you want to prefix all commands you want the working directory to
apply to with `wswd exec` in your `i3` config. e.g. I have

```
bindsym Mod4+Return exec $HOME/bin/wswd exec $HOME/bin/urxvt
bindsym Mod4+a exec $HOME/bin/wswd exec /usr/bin/dmenu_run
```

When I work on a project, I change to a new workspace and run `wswd set` in the
root of that project, e.g.

```bash
$ cd src/gonih.org/wswd
$ wswd set
$ vim main.go
```

Now, any new terminal I open (e.g. to run `go build`, `go get`, `go doc`…)
opens in `~/src/gonih.org/wswd`.

Once I'm done with my task, I close all windows, change workspaces and the
association is gone. If I ever get confused because a terminal opens in a weird
directory, I run `wswd unset`.

I also have a [systemd timer](https://wiki.archlinux.org/title/systemd/Timers)
set up to run `wswd clean` every hour, to garbage collect unused associations.
See `wswd.service` and `wswd.timer`.

# License

```
Copyright 2023 Axel Wagner

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
