# tfstatediff
Compare the tfstate before apply and the tfstate after apply, and comment the instance name and ip address in the github pull request.

:warning: **ci supports only drone and cloud supports only sakura cloud**

## How to use
`tfstatediff -old terraform.tfstate.old -new terraform.tfstate.new -conf ./example.yml`

## Settings
Please change repository owner adn name.
```
---
ci: drone
notifier:
  github:
    token: $GITHUB_TOKEN
    repository:
      owner: "ak1ra24"
      name: "tfstatediff"
```