# upguard-cli
A (Go) CLI tool to access [Upguard](https://www.upguard.com/) based on [Restish](www.rest.sh).
All supported commands, behaviors, features of restish are supported by upguard-cli.

## Features
upguard-cli, on top of restish features, supports also:
- pagination of upguard API respponse
- native swagger support

## What is Upguard
"UpGuard helps businesses manage cybersecurity risk. UpGuard's integrated risk platform combines third party security ratings, security assessment questionnaires, and threat intelligence capabilities to give businesses a full and comprehensive view of their risk surface."

## Usage
upguard-cli queries the Upguard API service to access information from the CyberRisk platform programmatically.

You can find or generate an API key to access this API in your CyberRisk Account Settings. The generate key should be export to the shell, for example 
```bash
export UPGUARD_CLI_UPGUARDKEY="xxxxx-xxx-xxx-xxxx-xxxx"
```

The base url for all public endpoints is https://cyber-risk.upguard.com/api/public . Please make sure that the URL is reachable (proxy/filters/firewalls).

- Subsidiaries Commands:
  - subsidiaries                        Get a list of subsidiaries
  - subsidiary-domain-details           Retrieve details for a domain
  - subsidiary-domain-update-labels     Assign labels to an domain
  - subsidiary-domains                  List subsidiary domains
  - subsidiary-ip-details               Retrieve details for an IP address
  - subsidiary-ip-update-labels         Assign labels to an IP
  - subsidiary-ips                      List subdiary ips
  - subsidiary-ranges                   List subsidiary ip ranges

- Domains Commands:
  - add-custom-domains                  Add custom domains
  - domain-details                      Retrieve details for a domain
  - domain-update-labels                Assign labels to a domain
  - domains                             Get a list of domains
  - remove-custom-domains               Remove custom domains

- Vendors Commands:
  - additional-evidence                 Retrieve (one or more) vendor additional evidence documents by id
  - additional-evidences-list           List vendor additional evidence instances
  - attachment                          Retrieve (one or more) vendor questionnaire attachments by id
  - attachments                         List vendor questionnaire attachments
  - document                            Retrieve (one or more) vendor documents by id
  - documents                           List vendor documents
  - monitorvendor                       Start monitoring a vendor
  - questionnaires                      List vendor questionnaires
  - questionnaires-v2                   List vendor questionnaires
  - unmonitorvendor                     Stop monitoring a vendor
  - vendor                              Get vendor details
  - vendor-domain-details               Retrieve details for a domain
  - vendor-domain-update-labels         Assign labels to an domain
  - vendor-domains                      List vendor domains
  - vendor-ip-details                   Retrieve details for an IP address
  - vendor-ip-update-labels             Assign labels to an IP
  - vendor-ips                          List vendor ips
  - vendor-ranges                       List vendor ip ranges
  - vendor-update-attributes            Assign or update the attributes for a vendor
  - vendor-update-labels                Assign labels to a vendor
  - vendor-update-tier                  Assign tier to a vendor
  - vendors                             List monitored vendors

- Bulk Commands:
  - bulk-deregister-hostnames           Deregister a list of hostnames
  - bulk-get-hostname-details           Get the details for a hostname
  - bulk-get-hostnames-stats            Get statistics around registered bulk hostnames
  - bulk-hostname-put-labels            Assign new labels to a hostname
  - bulk-list-hostnames                 List hostnames and their risks
  - bulk-register-hostnames             Register a list of hostnames to be scanned for risks

- Labels Commands:
  - labels                              Get the list of registered labels

- Ips Commands:
  - add-custom-ips                      Add custom ips
  - ip-details                          Retrieve details for an IP address
  - ip-update-labels                    Assign labels to an IP
  - ips                                 Get a list of ips
  - ranges                              Get a list of ip ranges
  - remove-custom-ips                   Remove custom ips

- Breaches Commands:
  - breached-identities                 Get a list of breached identities
  - identity-breach                     Get the details for an identity breach

- Risks Commands:
  - available-risks                     Get a list of available risks in the platform
  - available-risks-v2                  Get a list of available risks in the platform
  - org-risks-diff                      Get a list of risk changes for your account
  - risk                                Get details for a risk in the platform
  - risks                               Get a list of active risks for your account
  - vendor-questionnaire-risks          Get a list of questionnaire risks for one or more watched vendors or a specific questionnaire
  - vendor-questionnaire-risks-v2       (V2) Get a list of questionnaire risks for one or more watched vendors or a specific questionnaire
  - vendor-risks                        Get a list of active risks for a vendor
  - vendor-risks-diff                   Get a list of risk changes for a vendor
  - vendors-risks-diff                  Get a list of risk changes for monitored vendors

- Reports Commands:
  - custom-reports-list                 Get the list of custom report templates defined for the account
  - queue-report                        Queue a report export
  - report-status                       Get the status of an exported report

- Dataleaks Commands:
  - dataleaks-disclosures               Get a list of disclosures
  - dataleaks-disclosures-update-status Update the status of a disclosure

- Vulnerabilities Commands:
  - org-vulnerabilities                 List potential vulnerabilities of your domains & IPs
  - vendor-vulnerabilities              List potential vulnerabilities of a vendor

- Notifications Commands:
  - get-notifications                   Get a list of notifications for your organization.

- Webhooks Commands:
  - create-webhook                      Create a new webhook
  - delete-webhook                      Delete a webhook
  - list-webhooks                       List webhooks
  - sample-webhook                      Webhook example data
  - webhooks-notification-types         Webhook notification types

- Typosquat Commands:
  - list-typosquat-domains              List typosquat domains
  - typosquat-details                   Retrieve typosquat details for a domain.

- Organisation Commands:
  - get-organisation-v1                 Get the current organisation




Use "upguard-cli [command] --help" for more information about a command.

## Example

```bash 
bash-5.2$ ./upguard-cli get-organisation-v1
HTTP/2.0 200 OK
Access-Control-Allow-Origin: https://cyber-risk.upguard.com
Access-Control-Expose-Headers: Authorization, Authorization-Expires, Authorization-Orgid, Authorization-Updated, Content-Disposition
Alt-Svc: h3=":443"; ma=2592000,h3-29=":443"; ma=2592000
Cache-Control: no-store
Content-Length: 210
Content-Type: application/json
Date: Sun, 07 Apr 2024 08:52:01 GMT
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
Vary: Accept-Encoding
Via: 1.1 google
X-Content-Type-Options: nosniff
X-Frame-Options: sameorigin

{
  automatedScore: 880
  categoryScores: {
    brandProtection: 947
    emailSecurity: 898
    networkSecurity: 934
    phishing: 950
    websiteSecurity: 849
  }
  id: 1920
  name: "MyOrg"
  primary_hostname: "my.org"
}
```

## Configuration
upguard-cli will automatically create its configuration file in one of the following location:

|OS|Path|
| :------------------- | :---------- | 
|Mac|	~/Library/Application Support/upguard-cli/apis.json|
|Windows|	%AppData%\upguard-cli\apis.json|
|Linux|	~/.config/upguard-cli/apis.json|

the following environment should be set: *UPGUARD_CLI_UPGUARDKEY*

> [!WARNING]
> You should not modify the configuration files by hand ... 