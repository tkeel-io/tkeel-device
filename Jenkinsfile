pipeline {
  agent {
    node {
      label 'nodejs'
    }
  }
    parameters {
        string(name:'DIY',defaultValue: 'no',description:'手动档/自动档 = yes/no')
        string(name:'GITHUB_ACCOUNT',defaultValue: 'lunz1207',description:'测试/正式仓库 = tkeel-io/lunz1207')
        string(name:'APP_VERSION',defaultValue: '0.0.0-testing',description:'组件 image 版本')
        string(name:'CHART_VERSION',defaultValue: '0.0.0-testing',description:'组件 chart 版本')
    }

    environment {
        /*
        相关信息
        */
        APP_NAME = 'tkeel-device'
        REGISTRY = 'docker.io'
        DOCKERHUB_NAMESPACE = 'tkeelio'
        /*
        凭证
        */
        DOCKER_CREDENTIAL_ID = 'dockerhub'
        GITHUB_CREDENTIAL_ID = 'github'
        KUBECONFIG_CREDENTIAL_ID = 'kubeconfig'
    }

    stages {
        stage ('checkout scm') {
            steps {
                checkout(scm)
            }
        }

        stage('set env'){
            steps {
                container ('nodejs'){
                    withCredentials([usernamePassword(credentialsId: "$GITHUB_CREDENTIAL_ID", passwordVariable: 'GIT_PASSWORD', usernameVariable: 'GIT_USERNAME')]) {
                        script {
                            /*
                            1.自动触发
                            2.若 tag 和 commmit id 匹配，则以 tag 为 docker image tag 和 helm chart verison,推送至正式仓库
                            2.若 tag 和 commmit id 不匹配，则以 commmit id 为 docker image tag,以当前时间为 helm chart verison,推送至测试仓库
                            */
                            sh 'git fetch --all --tags'
                            env.HELM_CHART_VERSION = "${sh(script:'git describe --abbrev=0 --tags',returnStdout:true)}"
                            if (env.HELM_CHART_VERSION == "${sh(script:'git tag --contains `git rev-parse HEAD`',returnStdout:true)}" ){
                                env.DOCKER_IMAGE_TAG = env.HELM_CHART_VERSION
                                env.GITHUB_ORG = 'tkeel-io'
                            }else{
                                env.DOCKER_IMAGE_TAG = "${sh(script:'git rev-parse --short HEAD',returnStdout:true)}"
                                env.HELM_CHART_VERSION = "${sh(script:'date -d "+8 hour" "+%m.%d.%H%M%S"',returnStdout:true)}"
                                env.GITHUB_ORG = 'lunz1207'
                            }
                            /*
                            1.手动触发
                            2.以传入的参数作为 docker image tag 和 chart verison
                            */
                            if (params.DIY == 'yes'){
                                sh 'echo "do it yourself"'
                                env.GITHUB_ORG = params.GITHUB_ACCOUNT
                                env.DOCKER_IMAGE_TAG = params.APP_VERSION
                                env.HELM_CHART_VERSION = params.CHART_VERSION
                            }
                            sh 'echo Docker image tag:$DOCKER_IMAGE_TAG'
                            sh 'echo Helm chart version:$HELM_CHART_VERSION'
                            sh 'echo Helm chart repo:$GITHUB_ORG'
                            sh 'echo Branch:$GIT_BRANCH'
                        }
                    }
                }
            }
        }
 
        stage('build & push chart') {
            environment {
                CHART_REPO_PATH = '/home/jenkins/agent/workspace/helm-charts'
            }
            steps {
                container ('nodejs') {
                    sh 'helm3 package chart/$APP_NAME --app-version=$DOCKER_IMAGE_TAG --version=$HELM_CHART_VERSION'
                    withCredentials([usernamePassword(credentialsId: "$GITHUB_CREDENTIAL_ID", passwordVariable: 'GIT_PASSWORD', usernameVariable: 'GIT_USERNAME')]) {
                        sh 'git config --global user.email "lunz1207@yunify.com"'
                        sh 'git config --global user.name "lunz1207"'
                        sh 'mkdir -p $CHART_REPO_PATH'
                        sh 'git clone https://$GIT_USERNAME:$GIT_PASSWORD@github.com/$GITHUB_ORG/helm-charts.git $CHART_REPO_PATH'
                        sh 'mv ./$APP_NAME-*.tgz $CHART_REPO_PATH/'
                        sh 'cd $CHART_REPO_PATH && helm3 repo index . --url=https://$GITHUB_ORG.github.io/helm-charts'
                        sh 'cd $CHART_REPO_PATH && git add . '
                        sh 'cd $CHART_REPO_PATH && git commit -m "feat:update chart"'
                        sh 'cd $CHART_REPO_PATH && git push https://$GIT_USERNAME:$GIT_PASSWORD@github.com/$GITHUB_ORG/helm-charts.git'
                    }
                }
            }
        }

        stage ('build & push image') {
            steps {
                container ('nodejs') {
                    sh 'docker build -t $REGISTRY/$DOCKERHUB_NAMESPACE/$APP_NAME:$DOCKER_IMAGE_TAG .'
                    withCredentials([usernamePassword(passwordVariable : 'DOCKER_PASSWORD' ,usernameVariable : 'DOCKER_USERNAME' ,credentialsId : "$DOCKER_CREDENTIAL_ID" ,)]) {
                        sh 'echo "$DOCKER_PASSWORD" | docker login $REGISTRY -u "$DOCKER_USERNAME" --password-stdin'
                        sh 'docker push $REGISTRY/$DOCKERHUB_NAMESPACE/$APP_NAME:$DOCKER_IMAGE_TAG'
                    }
                }
            }
        }

        stage('install or upgrade plugin') {
            steps {
                container ('nodejs') {
                    withCredentials([
                    kubeconfigFile(
                    credentialsId: env.KUBECONFIG_CREDENTIAL_ID,
                    variable: 'KUBECONFIG')
                    ]) {
                        sh 'wget -q https://raw.githubusercontent.com/tkeel-io/cli/master/install/install.sh -O - | /bin/bash'
                        sh 'tkeel admin login -p changeme'
                        sh 'sleep 1m'
                        sh 'tkeel plugin upgrade $GITHUB_ORG/$APP_NAME@$HELM_CHART_VERSION $APP_NAME'
                    }                
                }
            }
        }     

        stage ('testing') {
            environment {
                API_TESTS = '/home/jenkins/agent/workspace/api-tests'
            }
            steps {
                container ('nodejs') {
                  withCredentials([usernamePassword(credentialsId: "$GITHUB_CREDENTIAL_ID", passwordVariable: 'GIT_PASSWORD', usernameVariable: 'GIT_USERNAME')]) {
                        sh 'echo this is a test'
                        // sh 'mkdir -p $API_TESTS'
                        // sh 'git clone https://$GIT_USERNAME:$GIT_PASSWORD@github.com/tkeel-io/tests.git $API_TESTS'
                        // sh 'cd $API_TESTS && npm install'
                        // sh 'cd $API_TESTS && npm run test /tests/device'
                    }
                }
            }
        }
    }

    post { 
        failure { 
            mail(
                to: 'lunzhou@yunify.com', 
                cc: '', 
                subject: 'tkeel-device pipeline is failure', 
                body:'failure'
            )
        }
        success { 
            mail(
                to: 'lunzhou@yunify.com', 
                cc: '', 
                subject: 'tkeel-device pipeline is success', 
                body:'success'
            )
        }
    }
}