FROM jenkinsci/jenkins:2.49

USER root
RUN apt-get update \
&& apt-get install -y --no-install-recommends build-essential \
&& rm -rf /var/lib/apt/lists/*
USER jenkins

ADD config/id_rsa /var/jenkins_home/.ssh/id_rsa

RUN /usr/local/bin/install-plugins.sh golang ws-cleanup timestamper slack \
    test-results-analyzer git

# XXX: We unset the Entrypoint so that specs can run arbitrary commands (such
# as `bash`) to initialize the container. This is necessary to properly write
# files into JENKINS_HOME.
Entrypoint []
