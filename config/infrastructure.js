const {Machine} = require("@quilt/quilt");

function MachineDeployer(nWorker) {
    this.nWorker = nWorker;
}

MachineDeployer.prototype.deploy = function(deployment) {
    var baseMachine = new Machine({
        provider: "Amazon",
        region: "us-west-1",
        sshKeys: ["ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCxMuzNUdKJREFgUkSpD0OPjtgDtbDvHQLDxgqnTrZpSvTw5r8XDd+AFS6eVibBfYv1u+geNF3IEkpOklDlII37DzhW7wzlRB0SmjUtODxL5hf9hKoScDpvXG3RBD6PBCyOHA5IJBTqPGpIZUMmOlXDYZA1KLaKQs6GByg7QMp6z1/gLCgcQygTDdiTfESgVMwR1uSQ5MRjBaL7vcVfrKExyCLxito77lpWFMARGG9W1wTWnmcPrzYR7cLzhzUClakazNJmfso/b4Y5m+pNH2dLZdJ/eieLtSEsBDSP8X0GYpmTyFabZycSXZFYP+wBkrUTmgIh9LQ56U1lvA4UlxHJ"],
    });

    deployment.deploy(baseMachine.asMaster());

    var worker = baseMachine.asWorker();
    worker.preemptible = true;
    deployment.deploy(worker.replicate(this.nWorker));
}

// We will have three worker machines.
module.exports = new MachineDeployer(3);
