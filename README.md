Go Reload
==========

Go Reload doesn't do preprocessing. Meaning u must handle SASS, Coffee, LESS yoruself or any pre-processing lib. Go Reload only reload the page for you when you made change and that's it. It doesn't do anything special other than that.

1. Clone this repo

2. Run `goreload`

```
./goreload -p port_to_run_go_reload_on_default_51203 -d /path/to/project/folder/to/look

```
* -p The port to run on, should > 1024 to run without admin perm.
* -d The directory to watch for the changes. 

3. Include script tag: http://127.0.0.1:8080/goreload.js

4. Edit the file and see browser auto reload

5. Watch our video: http://youtu.be/OmbNpV4c6vs

How it works
============

It utilize ~~https://github.com/alandipert/fswatch~~ https://github.com/howeyc/fsnotify/
Whenever you makes a change to any files inside the directory `goreload` watched, goreload notices the timestamp for that change, store it and refresh the page.
