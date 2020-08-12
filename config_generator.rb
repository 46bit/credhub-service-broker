#!/usr/bin/env ruby

require 'json'

config = {
  log_level: "INFO",
  port: "8080",

  basic_auth_username: ENV["BASIC_AUTH_USERNAME"],
  basic_auth_password: ENV["BASIC_AUTH_PASSWORD"],

  credhub_url: ENV["CREDHUB_URL"],
  credhub_prefix: ENV["CREDHUB_PREFIX"],
  credhub_uaa_client_id: ENV["CREDHUB_UAA_CLIENT_ID"],
  credhub_uaa_client_secret: ENV["CREDHUB_UAA_CLIENT_SECRET"],
  credhub_ca: ENV["CREDHUB_CA"],

  locket: {
    address: ENV["LOCKET_ADDRESS"],
    client_cert_file: "config_locket_client.crt",
    client_key_file: "config_locket_client.key",
    ca_cert_file: "config_locket_ca.crt",
  },

  catalog: {
    services: [
      {
        id: "55a7382a-906a-4011-a1fd-6b440652cda4",
        name: "secrets",
        description: "Secrets management backed by Credhub",
        bindable: true,
        plan_updateable: false,
        metadata: {
          displayName: "Credhub Secrets management",
          shareable: true
        },
        plans: [
          {
            id: "24efab31-8cbd-47c0-8513-a9345f3c512b",
            name: "simple",
            description: "Straightforward secrets management with `cf update-service`",
            free: true,
            metadata: {
              displayName: "Simple",
              AdditionalMetadata: {
                encrypted: true
              }
            }
          }
        ],
        tags: ["secrets", "credhub"]
      }
    ]
  }
}

File.write "config.json", JSON.pretty_generate(config)

File.write config[:locket][:client_cert_file], ENV["LOCKET_CLIENT_CERT"]
File.write config[:locket][:client_key_file], ENV["LOCKET_CLIENT_KEY"]
File.write config[:locket][:ca_cert_file], ENV["LOCKET_CA_CERT"]
