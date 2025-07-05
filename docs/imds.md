# OCI IMDSv2 Metadata Summary – `sample node`

ref: https://docs.oracle.com/en-us/iaas/Content/Compute/Tasks/gettingmetadata.htm

This document captures and explains the key instance metadata retrieved via IMDSv2 on a compute instance running in Oracle Cloud Infrastructure (OCI). Data was fetched using the following commands:

```bash
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/instance/
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/vnics/
curl -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/identity/
```

---

## 1. Instance Metadata (`/opc/v2/instance/`)

### Sample Output

```json
{
  "agentConfig": {
    "allPluginsDisabled": false,
    "managementDisabled": false,
    "monitoringDisabled": false,
    "pluginsConfig": [
      {
        "desiredState": "ENABLED",
        "name": "Compute HPC RDMA Authentication"
      },
      {
        "desiredState": "ENABLED",
        "name": "Compute HPC RDMA Auto-Configuration"
      },
      {
        "desiredState": "ENABLED",
        "name": "Compute RDMA GPU Monitoring"
      }
    ]
  },
  "availabilityDomain": "kWVD:US-rekharoy-1-AD-3",
  "canonicalRegionName": "us-rekharoy-1",
  "compartmentId": "ocid1.compartment.oc1..rekharoy",
  "definedTags": {
    "Oracle-Tag": {
      "CreatedBy": "ocid1.instance.oc1.us-chicago-1.rekharoy",
      "CreatedOn": "2025-06-12T02:32:22.888Z"
    }
  },
  "displayName": "BIO-2334XLG08T",
  "faultDomain": "FAULT-DOMAIN-1",
  "freeformTags": {
    "hostSerial": "2334XLG08T"
  },
  "hostname": "bio-2334xlg08t",
  "id": "ocid1.instance.oc1.us-chicago-1.rekharoy",
  "image": "ocid1.image.oc1.us-chicago-1.rekharoy",
  "metadata": {
    "ssh_authorized_keys": "rekharoy"
  },
  "ociAdName": "us-rekharoy-1-ad-1",
  "region": "us-rekharoy-1",
  "regionInfo": {
    "realmDomainComponent": "oraclecloud.com",
    "realmKey": "oc1",
    "regionIdentifier": "us-rekharoy-1",
    "regionKey": "rekharoy"
  },
  "shape": "BM.GPU.H100.8",
  "shapeConfig": {
    "maxVnicAttachments": 256,
    "memoryInGBs": 2048.0,
    "networkingBandwidthInGbps": 100.0,
    "ocpus": 112.0
  },
  "state": "Running",
  "tenantId": "ocid1.tenancy.oc1..rekharoy",
  "timeCreated": 1749695543362
}
```

---

## 2. VNIC Metadata (`/opc/v2/vnics/`)

### Sample Output

```json
[
  {
    "macAddr": "rekharoy:3f:d2:b3:0b:0c",
    "nicIndex": 0,
    "privateIp": "10.0.11.179",
    "subnetCidrBlock": "10.0.8.0/21",
    "virtualRouterIp": "10.0.8.1",
    "vlanTag": 0,
    "vnicId": "ocid1.vnic.oc1.us-chicago-1.rekharoy"
  }
]
```

---

## 3. Identity Metadata (`/opc/v2/identity/`)

The identity endpoint provides cryptographic material for instance principals. Key components include:

* `cert.pem`: Instance certificate
* `intermediate.pem`: Intermediate CA certificate
* `key.pem`: Private key for identity
* `fingerprint`: Certificate fingerprint (short identifier, useful for validation)
* `tenancyId`: OCID of the tenancy this instance belongs to

**Note**: The full output is very long. Below is a truncated sample:

```json
{
  "cert.pem": "-----BEGIN CERTIFICATE-----\n...<rekharoy>...\n-----END CERTIFICATE-----\n",
  "intermediate.pem": "-----BEGIN CERTIFICATE-----\n...<rekharoy>...\n-----END CERTIFICATE-----\n",
  "key.pem": "-----BEGIN RSA PRIVATE KEY-----\n...<rekharoy>...\n-----END RSA PRIVATE KEY-----\n",
  "fingerprint": "test:fingerprint:12345",
  "tenancyId": "ocid1.tenancy.oc1..rekharoy"
}
```

> 🔒 **Important**: Do not expose or log the full `key.pem`. It grants identity-based authentication for the instance.

---

## Best Practices

* Use IMDSv2 only; disable IMDSv1 via Console or API
* Always include `Authorization: Bearer Oracle` header
* Cache metadata values to avoid hitting rate limits
* Secure `key.pem` if used (rotate regularly and never share)

---
