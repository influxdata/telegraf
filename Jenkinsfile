properties(
	[
		buildDiscarder(logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '', numToKeepStr: '10')),
		pipelineTriggers([pollSCM('0 H(5-6) * * *')])
	]
)

pipeline
{
	agent { node { label 'linux && development' } }
	
	stages
	{
		stage('Checkout')
		{
			steps
			{
				cleanWs()
				dir('src/github.com/influxdata/telegraf') 
				{
					checkout scm
				}
				sh 'mkdir ./release'
			}
		}
		
		stage('Prepare deps') 
		{
			steps
			{
				sh '''
					export GOROOT=/usr/local/go
					export PATH=$PATH:$GOROOT/bin
					export GOPATH=${WORKSPACE}
					cd ./src/github.com/influxdata/telegraf
					make -f Makefile_jenkins deps'''
			}
		}
		
		stage('Build-Windows') 
		{
			steps
			{
				sh '''
					export GOROOT=/usr/local/go
					export PATH=$PATH:$GOROOT/bin
					export GOPATH=${WORKSPACE}
					workspace=`pwd`
					cd ./src/github.com/influxdata/telegraf
					export GIT_SHORT="$(git rev-parse --short HEAD)"
					export BUILD_DATE=$(date +"%Y%m%d")
					make -f Makefile_jenkins build-windows
					cd $workspace
					mv ./src/github.com/influxdata/telegraf/telegraf.exe ./release/telegraf-${BUILD_DATE}-${GIT_SHORT}.exe'''
			}
		}
		stage('Build-Linux') 
		{
			steps
			{
				sh '''
					export GOROOT=/usr/local/go
					export PATH=$PATH:$GOROOT/bin
					export GOPATH=${WORKSPACE}
					workspace=`pwd`
					cd ./src/github.com/influxdata/telegraf
					export GIT_SHORT="$(git rev-parse --short HEAD)"
					export BUILD_DATE=$(date +"%Y%m%d")
					make -f Makefile_jenkins build-linux
					cd $workspace
					mv ./src/github.com/influxdata/telegraf/telegraf ./release/telegraf-${BUILD_DATE}-${GIT_SHORT}'''
			}
		}
		
		stage('Build-Linux-arm') 
		{
			steps
			{
				sh '''
					export GOROOT=/usr/local/go
					export PATH=$PATH:$GOROOT/bin
					export GOPATH=${WORKSPACE}
					workspace=`pwd`
					cd ./src/github.com/influxdata/telegraf
					export GIT_SHORT="$(git rev-parse --short HEAD)"
					export BUILD_DATE=$(date +"%Y%m%d")
					make -f Makefile_jenkins linux_arm-build
					cd $workspace
					mv ./src/github.com/influxdata/telegraf/telegraf.arm ./release/telegraf-${BUILD_DATE}-${GIT_SHORT}.arm'''
			}
		}
		
		stage('Archive')
		{
			steps
			{
				archiveArtifacts artifacts: 'release/telegraf*', onlyIfSuccessful: true
			}
		}
		stage('CleanUp')
		{
			steps
			{
				cleanWs()
				notifySuccessful()
			}
		}
	}
	post 
	{ 
        failure { 
            notifyFailed()
        }
		success { 
            notifySuccessful()
        }
		unstable { 
            notifyFailed()
        }
    }
}

def notifySuccessful() {
	echo 'Sending e-mail'
	mail (to: 'notifier@manobit.com',
         subject: "Job '${env.JOB_NAME}' (${env.BUILD_NUMBER}) success build",
         body: "Please go to ${env.BUILD_URL}.");
}

def notifyFailed() {
	echo 'Sending e-mail'
	mail (to: 'notifier@manobit.com',
         subject: "Job '${env.JOB_NAME}' (${env.BUILD_NUMBER}) failure",
         body: "Please go to ${env.BUILD_URL}.");
}