# Provider

- [1. Overview](#1-overview)
- [2. Technical Details](#2-technical-details)
    - [2.1 Types of provider integration](#21-types-of-provider-integration)
      - [2.1.1 Out-of-process](#211-out-of-process)
      - [2.1.2 In-process](#212-in-process)
    - [2.2 Provider lifecycle](#22-provider-lifecycle)
      - [2.2.1 Registration](#221-registration)
      - [2.2.2 Initialization](#222-initialization)
      - [2.2.3 Source mapping](#223-source-mapping)
      - [2.2.4 Usage preparation](#224-usage-preparation)
      - [2.2.5 Opening sources](#225-opening-sources)

## 1. Overview

Incredible does not manage/store secrets itself.
Instead, it relies upon a third-party provider like Bitwarden, Last Pass or Azure Key Vault to provide a secure way of
persisting and retrieving sensitive data.

## 2. Technical Details

The provider integration is defined by the [provider interface](/pkg/provider/provider.go).

### 2.1 Types of provider integration

Depending on the technical details of and the interfaces provided by a provider's backend, the provider interface
implementation can be categorized into two types.

#### 2.1.1 Out-of-process

This kind of integration relies on an external program, which will be invoked by `incredible`.

```text
+------------+                   +--------------------+
| incredible | ---in-process---> | provider interface |
+------------+                   +--------------------+
                                           /\
                                           ||
      out-of-process std I/O communication || 
                                           ||
                                           \/
                                    +--------------+
                                    | provider cli |
                                    +--------------+       
```

In most cases, the external program will be the provider's official CLI client.
This comes with the benefit of not having to deal with any authentication information within `incredible` (like vault
master passwords).
Additionally, the CLI client can be updated independently (assuming the command line interface remains compatible).

If there is an official and maintained CLI client available for the three major platforms (Linux, Windows, macOS), this
type of provider integration should be preferred.

#### 2.1.2 In-process

An in-process provider integration contains the functionality to interact with the provider's API within the
main `incredible`-executable.
For that purpose, the `Provider` interface implementation might rely on an official or third-party library.

### 2.2 Provider lifecycle

A provider normally passes through the stages outlined in the following table.

| #   | Phase             | Description                                                                                                                                                             | Interface method          |
|-----|-------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------|
| 0   | Registration      | The provider adds itself to the `provider.Providers` slice.                                                                                                             | -                         |
| 1   | Initialization    | The provider performs basic initilization tasks or returns `ErrProviderUnavailable` if necessary preconditions aren't met.                                              | `Provider.Initialize`     |
| 2   | Source mapping    | The `.SupportSource` method is invoked to check if the provider supports a given source.                                                                                | `Provider.SupportsSource` |                                                                                                                                                                  
| 3   | Usage preparation | If at least one declared asset relies on the provider, the method is invoked and the provider should prepare itself for usage (e.g. unlocking vault, performing login). | `Provider.PrepareUsage`   |
| 4   | Opening sources   | The `.Open` method is invoked for each supported source.                                                                                                                | `Provider.Open`           |

#### 2.2.1 Registration
In order to register a provider and make it usable by `incredible`, an implementation of `Provider` has to be added to the `provider.Providers` slice.

This can be done within a package's `init` method.
In order to load this package and execute the `init` method, a side effect import should be added to [cmd/provider_imports.go](/cmd/provider_imports.go).

During registration no further actions should be performed by the provider.

#### 2.2.2 Initialization
The `.Initialize` method is invoked on each registered provider instance.
During the invocation the provider should check if all it's requirements are met (e.g. dependencies are present).

If the provider detects that at least one requirement is unmet, a `ErrProviderUnvailable` error should be returned.
`incredible` will then ignore the affected provider during further execution.
The provider may wrap the `ErrProviderUnvailable` within a more detailed error.
If any other error is returned, the execution is halted. 

It is also appropriate to perform other minor initialization tasks within the `.Initialize` method.
At this point it is still unclear if the provider is even required by at least one declared source,
which is why the provider should refrain from doing resource- or time-consuming tasks.
Especially there must be no action requiring user interaction.

#### 2.2.3 Source mapping
`incredible` will invoke the provider's `.SupportSource` method for each source defined within a given `incredible.yml` file.

The provider is expected to return `true`, if it should be able to successfully open the source that has been passed along during method invocation.
If the provider doesn't understand or support the given source specification, `false` should be returned.

The checks performed at this point should be limited to the information provided to the method by its parameters.
A result of `true` should only indicate that the provider is technically able to understand and handle the provided source specification.

#### 2.2.4 Usage preparation
Each provider that has returned `true` at least once during [source mapping](#223-source-mapping), is now being prepared for usage.
During this preparation, the `.PrepareUsage` method is invoked.

As it is now clear, that the provider is indeed going to be used for at least one asset, remaining setup actions should now be performed. 

This includes but isn't limited to:

- Checking if the user is logged in
  - if not, performing the required login
- Unlocking the vault being used
- Obtaining a session

#### 2.2.5 Opening sources
The `.Open` method will be called for all sources supported by the provider.
It is the provider's task to determine if the given source represents a secret file or a secret value. 

If it is a file, an instance of `BinarySource` should be returned. If it is a secret value, an instance of `ValueSource` should be returned.


