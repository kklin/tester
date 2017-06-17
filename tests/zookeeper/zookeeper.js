const {createDeployment} = require("@quilt/quilt");
var zookeeper = require("@quilt/zookeeper");
var infrastructure = require("@quilt/tester/config/infrastructure")

var deployment = createDeployment();
deployment.deploy(infrastructure);
deployment.deploy(new zookeeper.Zookeeper(infrastructure.nWorker*2));
