// vim: et:ts=4:sw=4:ft=groovy
def jenkinsSlack(type){
    def jobInfo = "\n » ${env.JOB_NAME} ${env.BUILD_NUMBER} (<${env.BUILD_URL}|job>) (<${env.BUILD_URL}/console|console>)"
    if (type == 'start'){
        slackSend color: 'blue', message: "build started${jobInfo}"
    }
    if (type == 'finish'){
        def buildColor = currentBuild.result == null? "good": "warning"
        def buildStatus = currentBuild.result == null? "SUCCESS": currentBuild.result
        slackSend color: buildColor, message: "build finished - ${buildStatus}${jobInfo}"
    }
}

def gitTags(commit) {
    sh("git tag --contains ${commit} > GIT_TAGS")
    def gitTags = readFile('GIT_TAGS').trim()
    sh('rm -f GIT_TAGS')
    if (gitTags == '') {
        return []
    }
    return gitTags.tokenize('\n')
}

def gitCommit() {
    sh('git rev-parse HEAD > GIT_COMMIT')
    def gitCommit = readFile('GIT_COMMIT').trim()
    sh('rm -f GIT_COMMIT')
    return gitCommit
}

def gitMasterBranchCommit() {
    sh('git rev-parse origin/master > GIT_MASTER_COMMIT')
    def gitCommit = readFile('GIT_MASTER_COMMIT').trim()
    sh('rm -f GIT_MASTER_COMMIT')
    return gitCommit
}

def onMasterBranch(){
    return gitCommit() == gitMasterBranchCommit()
}

def imageTags(){
    def gitTags = gitTags(gitCommit())
    if (gitTags == []) {
        return ["canary"]
    } else {
        return gitTags + ["latest"]
    }
}

node('docker'){
    catchError {
        def imageName = 'simonswine/slingshot'
        def imageTag = 'jenkins-build'

        jenkinsSlack('start')

        stage 'Checkout source code'
        checkout scm

        stage 'Build and test slingshot'
        sh "make docker"

        stage 'Full integration with vagrant and ansible'
        // prepare temp home directory
        def homePath = "${pwd(tmp: true)}/home"
        sh "rm -rf ${homePath}/.slingshot ${homePath}/.kube ${homePath}/.ssh"
        sh "mkdir -p ${homePath} ${homePath}/.vagrant.d ${homePath}/.kube"
        sh "cp ~/.vagrant.d/insecure_private_key ${homePath}/.vagrant.d/"
        env.HOME = homePath

        sh "./_build/slingshot-linux-amd64 cluster create -I simonswine/slingshot-ip-vagrant-coreos -C simonswine/slingshot-cp-ansible-k8s-contrib cluster1"

        // copy kubectl config over
        sh "ssh -o \"UserKnownHostsFile /dev/null\" -o \"StrictHostKeyChecking no\" -i ~/.vagrant.d/insecure_private_key core@10.251.0.10 cat /etc/kubernetes/kubectl.kubeconfig > ~/.kube/config"

        // get node status
        sh "kubectl get nodes"

        // schedule a pod
        sh "kubectl run --attach --image busybox --restart=Never testpod ping -- -c 10 8.8.4.4"


    }
    stage 'Cleanup virtual box instances'
    sh """for machine in `VBoxManage list vms | grep slingshot | awk '{print \$2}'`; do
        VBoxManage controlvm \${machine} poweroff || true
        VBoxManage unregistervm \${machine} --delete
    done"""
    jenkinsSlack('finish')
    step([$class: 'Mailer', recipients: 'christian@jetstack.io', notifyEveryUnstableBuild: true])
}

