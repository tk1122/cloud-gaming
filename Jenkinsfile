pipeline {
    environment {
        DEPLOY = "${env.BRANCH_NAME == "master" ? "true" : "false"}"
        REGISTRY = 'slark1122/cloud-gaming'
        REGISTRY_CREDENTIAL = 'dockerhub-slark1122'
    }
    agent {
        kubernetes {
            defaultContainer 'jnlp'
            yamlFile 'jenkins-slave.yaml'
        }
    }
    stages {
        stage('Docker Build') {
            when {
                environment name: 'DEPLOY', value: 'true'
            }
            steps {
                container('docker') {
                    sh "docker build -t ${REGISTRY}:latest ."
                }
            }
        }
        stage('Docker Publish') {
            when {
                environment name: 'DEPLOY', value: 'true'
            }
            steps {
                container('docker') {
                    withDockerRegistry([credentialsId: "${REGISTRY_CREDENTIAL}", url: ""]) {
                        sh "docker push ${REGISTRY}:latest"
                    }
                }
            }
        }
        stage('Kubernetes Deploy') {
            when {
                environment name: 'DEPLOY', value: 'true'
            }
            steps {
                container('helm') {
                    sh "helm upgrade --install --force cloud-gaming ./k8s/cloud-gaming"
                }
            }
        }
    }
}
