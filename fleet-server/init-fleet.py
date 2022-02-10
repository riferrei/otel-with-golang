import os
import requests
from requests.auth import HTTPBasicAuth

es_url = os.getenv("ELASTICSEARCH_HOST")
kb_url = os.getenv("KIBANA_HOST")
es_user = os.getenv("ELASTICSEARCH_USERNAME")
es_pass = os.getenv("ELASTICSEARCH_PASSWORD")
basic_auth = HTTPBasicAuth(es_user, es_pass)

default_headers = {
    "Content-Type": "application/json",
    "kbn-xsrf": "true"
}

wait = True
while wait:
    try:
        res = requests.get(
            url=es_url,
            auth=basic_auth,
            headers=default_headers
        )
        if res.status_code == 200:
            wait = False
    except:
        print("Waiting for Elasticsearch")

token_url = "{}/_security/oauth2/token".format(es_url)
data_json = {"grant_type": "client_credentials"}
res = requests.post(
    url=token_url,
    auth=basic_auth,
    json=data_json,
    headers=default_headers
)
es_access_token = res.json()['access_token']

service_toke_url = "{}/_security/service/elastic/fleet-server/credential/token".format(es_url)
data_json = {"grant_type": "client_credentials"}
default_headers['Authorization'] = "Bearer {}".format(es_access_token)

res = requests.post(
    url=service_toke_url,
    headers=default_headers
)
fleet_service_token = res.json()['token']['value']

print("ES_ACCESS_TOKEN={}".format(es_access_token))
print("FLEET_SERVICE_TOKEN={}".format(fleet_service_token))

with open("/out/environment", "w") as f:
    f.write("export ES_ACCESS_TOKEN={}\n".format(es_access_token))
    f.write("export FLEET_SERVER_SERVICE_TOKEN={}\n".format(fleet_service_token))
    f.write("export KIBANA_FLEET_SERVICE_TOKEN={}\n".format(fleet_service_token))
