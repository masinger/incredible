# Incredible [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
Incredible is a CLI tool that loads secrets (passwords, tokens, files) from secure sources and maps them to usable environment variables.

The goal of this tool is to avoid
- having to add sensitive values into your Bash profile
- permanently storing confidential files on the system
- manually copying secrets from your password manager into your bash

Just configure your desired environment variables (see ["The solution"](#the-solution)) and run your program with

```shell
incredible <MY_PROGRAM>
```
and it will be able to access them.

## Supported sources

| Source                                                                  | Requirements                                              | Value Support | File Support | Help                     |
|-------------------------------------------------------------------------|-----------------------------------------------------------|---------------|--------------|--------------------------|
| [Bitwarden](https://bitwarden.com)                                      | [Bitwarden Cli](https://bitwarden.com/help/cli/)          | ✅             | ✅            | [Help](#bitwarden)       |
| [Azure Key Vault](https://azure.microsoft.com/en-us/products/key-vault) | [Azure Cli](https://learn.microsoft.com/en-us/cli/azure/) | ✅             | ❌            | [Help](#azure-key-vault) |


## How does it work?
`incredible` can best be understood as an interfaces to various credential/password/secret managers.

### The problem

Without `incredible`, one could easily come up with a script that automatically fetches a value from a password manager and exposes it using an environment variable, which can then be consumed by another tool or script. 

You know - something like this:
````shell
export MY_CLIENT_SECRET=$(generic-password-manager get secret super-secret-oidc-secret)
````
This scripted solution should do the job just fine.
Starting from that approach, a variety of problems and questions arise when it comes to working with multiple secrets, environments, team members and secret providers.

For example:
- How can I set multiple secrets, If my environment needs them? 
- What if one of them required secrets is a file?
- How do I clean up my environment after the variables and files are no longer needed?
- How can I switch between contexts requiring the same variables set to different values?

### The solution
`incredible` tries to simplify working with secrets as much as possible by providing a unified configuration file.
This configuration file specifies the required environment variables and where they can be obtained from. 
Once a program is started with `incredible <MY_EXEC>`, `incredible` will fetch all required values and cleans them up once your program exits.

The configuration file (usually named `incredible.yml`) looks like this:

````yaml
assets: # List of all assets to be loaded
  - src: # The src property tells incredible, where the secret's value can be obtained from
      azureKeyVaultSecret: {} # set this, when loading from an Azure Key Value
      bitwarden: {} # set this, when loading from a Bitwarden vault
    mappings:
      - env: 
          name: TARGET_VARIABLE # the name of the environment variable, that should hold the obtained secret
````

The JSON schema at https://raw.githubusercontent.com/masinger/incredible/main/schema.json can be used to enable code completion and validation (if supported by your editor).

See [Sources](#sources) for more information on how to configure the `src` property.


### Feature list
- Fetch secret values from various providers (as listed in "[Supported sources](#supported-sources)")
- Fetch files from various providers, storing them in temporary files
- Removal of created temporary files, once incredible exits
- Usage of context specific environment variables (the `incredible.yml` within the current or the first ancestor directory will be used)

## Sources

### Bitwarden

#### Requirements

- [Bitwarden Cli](https://bitwarden.com/help/cli/) must be installed

#### General

The Bitwarden source requires the identifier of the Bitwarden entry to be used.
The id can be obtained using the Bitwarden cli or by clicking on the entry (within the web
vault - https://vault.bitwarden.com) and inspecting the browser URL once the detail dialog opens.

#### Secret value from password

The following shows the minimal configuration required in order to read the password stored in entry `123-test-id` and
map it to the environment variable named `MY_SECRET_PW`.

```yaml
assets:
  - src:
      bitwarden:
        entry: 123-test-id # entry id
        field: password # optional, default: "password" 
    mappings:
      - env:
          name: MY_SECRET_PW
```

#### Secret value from password

If we set the property `field` to `"username"`, the entry's username will be used instead.

````yaml
assets:
  - src:
      bitwarden:
        entry: 123-test-id # entry id
        field: username
    mappings:
      - env:
          name: MY_SECRET_USER
````

#### Secret file from attachment
In order to load a secret file from a Bitwarden entry's attachment, the name of the attachment must be provided using the `attachment` property.
The environment variable `MY_SECRET_FILE_PATH` will then hold the filepath to the loaded file.

````yaml
assets:
  - src: 
      bitwarden: 
        entry: 12345-12345-12345-abcd-12345 # Entry id
        attachment: my-attachment.txt # file name as shown in Bitwarden
    mappings:
      - env: 
          name: MY_SECRET_FILE_PATH
````

### Azure Key Vault

#### Requirements

- [Azure Cli](https://learn.microsoft.com/en-us/cli/azure/) must be installed

#### Obtaining a value from a Azure Key Vault secret
Sourcing values from an Azure Key Vault Secret requires the secrets unique identifier,
which can be obtained by running

````shell
az keyvault secrete show --vault-name <NAME_OF_YOUR_KEYVAULT> --name <NAME_OF_YOUR_SECRET>
````

> **Warning**
> The returned entry id statically refers to the current secret's version.
> In order to always use the current version, omit the last path segment.

The following mapping will load the newest value of an Azure Key Vault secret and store it within the environment variable `MY_SECRET_VALUE`:
````yaml
assets:
  - src: 
      azureKeyVaultSecret:
        # Key Vault Secret's id (the lat version path segment is omitted, in order to always use the newest value)
        itemId: https://MY_TEST_KEYVAULT.vault.azure.net/secrets/MY_TEST_ENTRY
    mappings:
      - env: 
          name: MY_SECRET_VALUE
````



