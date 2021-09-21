# Introduction
Auto create initial server host for nginx and apache, create project root from git repo. Only for laravel project.

### Requirements
- Must run this application as `root` user
- Configure your server to have SSH connection between your server and git repository(gitlab, github, etc)
- Your server should have `wget` application installed
- Your server should have nginx running (if using nginx)
- Your server should have apache running (if using apache)