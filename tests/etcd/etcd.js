const {createDeployment} = require("@quilt/quilt");
var etcd = require("@quilt/etcd");
var infrastructure = require("@quilt/tester/config/infrastructure")

var deployment = createDeployment({});
deployment.deploy(infrastructure);
deployment.deploy(new etcd.Etcd(infrastructure.nWorker*2));
