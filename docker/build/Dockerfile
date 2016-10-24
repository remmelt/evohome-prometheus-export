FROM centos:latest
MAINTAINER Jonathan Turner <jt@jtnet.co.uk>

RUN yum install -y \
  golang \
  git

ADD scripts/go-build.sh /tmp/go-build.sh
RUN chmod 744 /tmp/go-build.sh
CMD [ "/tmp/go-build.sh" ]