const {createDeployment, Service, Container, LabelRule} = require("@quilt/quilt");
var infrastructure = require("../../config/infrastructure.js")

var deployment = new createDeployment({});
deployment.deploy(infrastructure);

var c = new Container("networkstatic/iperf3", ["-s"]);

// We want (nWorker - 1) machines with 1 container to test intermachine bandwidth.
// We want 1 machine with 2 containers to test intramachine bandwidth.
// Since inclusive placement is not implemented yet, guarantee that one machine
// has two iperf containers by exclusively placing one container on each machine,
// and then adding one more container to any machine.
var exclusive = new Service("iperf", c.replicate(infrastructure.nWorker));
exclusive.place(new LabelRule(true, exclusive));

var extra = new Service("iperfExtra", [c]);

exclusive.connect(5201, exclusive);
extra.connect(5201, exclusive);
exclusive.connect(5201, extra);
extra.connect(5201, exclusive);

deployment.deploy(exclusive);
deployment.deploy(extra);
