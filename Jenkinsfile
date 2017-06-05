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
		stage('Prepare')
		{
			steps
			{
				dir('src/github.com/influxdata/telegraf') 
				{
					checkout changelog: true, poll: true, scm: [$class: 'GitSCM', branches: [[name: '*/master']], doGenerateSubmoduleConfigurations: false, extensions: [[$class: 'PreBuildMerge', options: [fastForwardMode: 'FF', mergeRemote: 'origin', mergeTarget: 'HuaweiWinApi']], [$class: 'PreBuildMerge', options: [fastForwardMode: 'FF', mergeRemote: 'origin', mergeTarget: 'OpenHardwareMonitor']]], submoduleCfg: [], userRemoteConfigs: [[credentialsId: '5f43e7cc-565c-4d25-adb7-f1f70e87f206', url: 'https://github.com/marianob85/telegraf']]]
				}
			}
		}
		stage('Build') 
		{
			steps
			{
				sh '''
					export GOROOT=/usr/local/go
					export PATH=$PATH:$GOROOT/bin
					export GOPATH=${WORKSPACE}
					cd ./src/github.com/influxdata/telegraf
					export GIT_SHORT="$(git rev-parse --short HEAD)"
					export BUILD_DATE=$(date +"%Y%m%d")
					make windows
					mv telegraf.exe telegraf-${BUILD_DATE}-${GIT_SHORT}.exe'''
			}
		}
		stage('Archive')
		{
			steps
			{
				archiveArtifacts artifacts: 'src/github.com/influxdata/telegraf/telegraf*.exe', onlyIfSuccessful: true
			}
		}
		stage('CleanUp')
		{
			steps
			{
				deleteDir()
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