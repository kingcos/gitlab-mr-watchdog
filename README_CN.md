# gitlab-mr-watchdog

[![Build Status](https://travis-ci.org/kingcos/gitlab-mr-jira-issue-trigger.svg?branch=master)](https://travis-ci.org/kingcos/gitlab-mr-jira-issue-trigger) [![Go Report Card](https://goreportcard.com/badge/github.com/kingcos/gitlab-mr-jira-issue-trigger)](https://goreportcard.com/report/github.com/kingcos/gitlab-mr-jira-issue-trigger) [![GitHub license](https://img.shields.io/github/license/kingcos/gitlab-mr-jira-issue-trigger.svg)](https://github.com/kingcos/gitlab-mr-jira-issue-trigger/blob/master/LICENSE)

[English](README.md) | 中文

![What](What.png)

监控 GitLab Merge Request 状态，并及时通知。

> 如果你的团队使用企业微信，可以尝试使用该 [Go 脚本](https://gist.github.com/kingcos/86e399f455495a8bbbdc30a4b6e1d482) 作为 Shell 命令，以发送文本信息。

## 配置

```yml
GitLab:
  host: GITLAB_HOST_ADDRESS (REQUIRED)
  owner: GITLAB_PROJECT_OWNER (REQUIRED)
  project: GITLAB_PROJECT_NAME (REQUIRED)
  token: GITLAB_PUBLIC_USER_TOKEN (REQUIRED)

TimeOut:
  created: TIMEOUT_DURATION_SINCE_CREATED (MINUTES)
  updated: TIMEOUT_DURATION_SINCE_UPDATED (MINUTES)
  start: TIME_START_OF_A_DAY (hh:mm)
  end: TIME_END_OF_A_DAY (hh:mm)

Watchdog:
  duration: WATCHDOG_REFRESH_DURATION (MINUTES)
  action:
    sh: WATCHDOG_SHELL_ACTION
```
