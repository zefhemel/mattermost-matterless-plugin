module github.com/zefhemel/mattermost-matterless-plugin

go 1.16

require (
	github.com/go-git/go-git/v5 v5.3.0
	github.com/mattermost/mattermost-plugin-starter-template/build v0.0.0-20210429201558-f5cae51a20a8
	github.com/mattermost/mattermost-server/v5 v5.3.2-0.20210413123336-5f2c26dbda0a
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/zefhemel/matterless v0.0.0-04251888ab13090d5304a64972f93e6d8a9f505a
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5 // indirect
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/zefhemel/matterless => ../matterless
