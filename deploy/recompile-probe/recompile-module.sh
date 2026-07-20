mkdir -p rebuild-kindling-agent
cd rebuild-kindling-agent
curl -O https://github.com/harmonycloud/agent-libs/archive/refs/tags/kindling-falcolib-probe-29721003629.tar.gz
tar -zvxf kindling-falcolib-probe-29721003629.tar.gz -C agent-libs
cd agent-libs
docker pull kindlingproject/kernel-builder:latest
docker run -it -v /usr:/host/usr -v /lib/modules:/host/lib/modules -v $PWD:/source kindlingproject/kernel-builder:latest
cd ..
cat <<EOF > Dockerfile
FROM kindlingproject/kindling-agent:latest
COPY ./agent-libs/kindling-falcolib-probe/* /opt/kindling-extra-probe/
EOF
docker build -t kindlingproject/kindling-agent:latest-bymyself .