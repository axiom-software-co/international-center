

- IMPORTANT AXIOM RULE TO FOLLOW : outside of unit tests , consider mocks and stubs the worst architectural anti pattern ( stop metioning the fact that you are using real implmentations , this is explicit in the implementation , keep the naming professional )

- IMPORTANT AXIOM RULE TO FOLLOW : when you read a compact conversation summary , the first thing you should do is read all the markdown files in the project to get context .

- IMPORTANT AXIOM RULE TO FOLLOW : we should not create new files outside of our current folders and files structure ( this is deliberate ) . ensure you stop and ask for permission if you believe we should alter our folder and files structure .

- IMPORTANT AXIOM RULE TO FOLLOW : do not edit any markdown files without permission . ensure you stop and ask for permission and provide your reasoning

- IMPORTANT AXIOM RULE TO FOLLOW : Our database schemas should exactly match that in our [name]-SCHEMA.md markdown files

- IMPORTANT AXIOM RULE TO FOLLOW : consider stubs the worst anti-pattern ( we should use propper infrastructure and fix root issues as they arise ) ( we should not fall back to stubs in integration tests ) 

- IMPORTANT AXIOM RULE TO FOLLOW : only run integration tests when the entire podman compose development environment is up

- IMPORTANT AXIOM RULE TO FOLLOW : environment variables are defined only in the env files ( networking configuration , including ports  , should not be hardcoded , it should always come from the environment ) ( we should not have fallback networking configuration in implementation nor integration tests ) 


- note : we will not work on the public and admin website frontend at the moment ( once we confirm our infrastructure and api gateways CICD is working , we will consider working on the website )

- note : stop using the term 'medical-grade' ( this is implicit and there is no need to mention it in the naming system )

- note : we do not need cloudfalre CDN for now


[ important project related rules to follow ]

# development workflow

- test-driven development ( red phase , green phase , refactor phase ) ( tests drive and validate the design of our architecture ) ( creating new methods from refactoring should not require us to write new tests , this violates the contract-first testing principle ) ( you are allowed to modify the project and tests implementations as you see fit , since project and/or tests abstractions and/or implementations sometimes need to be updated ) ( when planning a new TDD cycle , provide a list of all the files you intend to edit and what you intend to do in each phase )

- use podman compose with containerd runtime to provision our local development environment ( resources and services ) ( do not mention docker ) ( 'podman-compose --env-file .env.development' should be used to manage podman containers in local development ) 

# architecture patterns

- cohesion over coupling
- best practices in our stack
- idiomatic go patterns

- handler , service , repository pattern, dapr-centric

- dependency inversion ( interfaces for variable concerns, dependency injection with interfaces ) ( concrete types for stable concerns ) 
- http server patterns 
- health check pattern 
- synchronous and asynchronous patterns were appropriate

- environment configuration pattern 

- warnings as errors
- singletons can only depend on singletons

- the result pattern is an anti pattern ( go has built in error handling ) 

## architecture layers

- the highest layer is the frontend
- the lowest layer is the infrastructure layer
- lower layers should not depend on nor be aware of higher layers

## database migrations

- automated development and testing migrations
- manual production migrations

## api gateways

- reverse proxy
- rate limiting
- security headers
- cors policies

- handle cross cutting concerns

### public apis gateway

- anonymous usage
- ip-based rate limiting ( 1000 req/min )
- public website origins
- azure cosmosdb backing store
- standard observability
- standard security

### admin apis gateway

- role-based access control
- user-based rate limiting ( 100 req/min )
- medical-grade audit loggging
- medical-grade security

## api domains

- domain shared kernels for public and admin apis
- vertical slice public and admin apis

## api versioning

-

## observability

- logging ( use structured logging , not concatenation ) ( each log should have key bits of information ( user ID , Correrlation ID , Request URL , APP Version , and so forth ) ( logs should be developer focused ) ( log levels : debug , information , waarning , error , critical ) ( not having 100% log delivery is okay )

- audits ( for medical-grade compliance ) ( losing any data is unacceptable ) ( store in grafana cloud loki )

## security

- security ( we must have fallback policies that get evaluated if no other policy is specifieid )

## testing

- arrange , act , assert
- contract-first testing ( testing interfaces/contracts rather than implementation details ) ( focused on preconditions/dependencies and postconditions/state-change )
- properly-based testing

- unit tests must use mock for dependencies to craete isolation
- integration tests must use real dependencies ( not mocks )
- end to end tests must use real dependencies and be done in aspire ( they should test the website for proper backend to frontend integration )

- all tests must have timeouts ( they should fail fast if something is wrong ) ( 5 seconds for unit tests ) ( 15 secnds for integration ) ( 30 seconds for end to end tests )

- do not use curl commmands or cli tools for testing ( test through our testing framework )

## version control , continuous integration , continuous delivery

- trunk based development for version control

# stack

## public website

- astro
- vue
- tailwind
- shadcn-vue
- vite
- vitest
- pinia
- bun runtime

- public api gateway for dynamic data

- do not use react
- do not do UI design testing

- cloudflare pages

## apis and events and public and admin api gateways

- golang
- golang-migrate

- sql

- dapr apis

- feature flags

## dapr stateless services

- orchestrator : dapr control plan container for production and local development 

- apis services : api service and dapr sidecar containers for production and local development 
- gateway services : gateway service and dapr sidecar containers for production and local development

- identity provider : authentik container for production and local development 
- telemetry collection : grafana agent and dapr sidecar containers for production and local development

- infrastructure configuration : yaml configuraation files for production and local development

### dapr sidecar middleware configurations

- name resolution : dapr built-in service invocation
- rate limiting : dapr ratelimit in for production ( dapr ratelimit for local development )
- cors : dapr cors in production ( dapr cors for local development )
- route checker : dapr route checker in production ( dapr route checker for local development )
- route alias : dapr route alias in production ( dapr route alias for local development )
- bearer : dapr bearer in production ( dapr bearer for local development )
- oauth2 : dapr oauth2 in production ( dapr oauth2 for local development )
- oauth2 client credentials : dapr oauth2 client credentials in production ( dapr oauth2 client credentials for local development )
- opa policies : dapr opa policies in production ( dapr opa policies for local development )

## managed stateful resources

- authenticated sessions and dapr bindings : upstash redis hosted for production ( redis container for local development )

- pub/sub : upstash redis hosted in production ( redis in container for local development ) 

- secret store : hashicorp vault cloud hosted for production ( hashicorp vault and vault data containers for local development )
- relational database ( includes configuration store and state store and identity storage and services storage ) : azure manged postgre hosted for production ( postgre container for local development ) 
- non-relational database : mongodb cloud hosted for production ( mongodb container for local development )
- file storage : azure blob storage hosted for production ( azurite blob storage emulator and azurite-data containers for local development https://github.com/Azure/Azurite )

## grafana cloud observability 

- objeservability : grafana , mimir , loki , tempo , pyroscope in container with their respective data volumes containers ( grafana-data , mimir-data , loki-data , tempo-data , pyroscope-data ) for production and local development 

## content delivery

- website hosting : cloudflare pages
- content delivery netowkr : cloudflare cdn

## persistent storage migrations

- go migrations runner

- development :
    approach : Aggressive - always migrate to latest
    rollback : Easy - can destroy and recreate
    safety_checks : Minimal
    automation : Full automation via Podman Compose
    
- staging:
    approach : Careful - migrate with validation
    rollback : Supported with confirmation
    safety_checks : Moderate validation
    automation : Pulumi orchestrated via GitHub Actions
    
- production :
    approach : Conservative - extensive validation
    rollback : Manual approval required
    safety_checks : Full validation and backup
    automation : Pulumi orchestrated with human approval

## deployment ( staging , production )

- state storage : azure blob storage ( pulumi state storage container ) ( manually created using the azure cli ) 
- github secrets : ( azure authentication , pulumi state storage , grafana cloud authentication , hashicorp vault cloud authentication ) ( manually created using the azure cli , cloudflare cli , grafana cli , hashicorp vault cli , github cli ) 

- github version control
- github container registry
- github actions workflows

- pulumi cli and go sdk

[ important general rules to follow ]

- ultrathink and deep dive and be comprehensive and be professional

- this is nixos and the admin password is 'unsecure'

- be causious of technical debt
- be causious of overengineering

- do not create documentation
- do not preserve legacy implementations
- do not implement experimental architectures not part of industry
- do not change the UI of our website unless explicitly asked to
- do not create simple ephimeral validation implementations in /temp/ directories to avoid disorder in source files
- do not creaate script files for projects ( this is an anti-pattern )
- do not stage and commit and push unless I explicitly ask you to

# Task Management Context Guidelines

- task descriptions must include WHY ( business reason , compliance requirement , architectural decision , so forth )
- task descriptions must include SCOPE ( which APIs , which components , which environments , so forth )
- task descriptions must include DEPENDENCIES ( what must complete first , integration points , so forth )
- task descriptions must include CONTEXT ( gateway architecture , medical compliance , environment specifics , so forth )

- critical : ensure you add proper context to your task list items , in the event context compression happens in the middle of a task , so you have a better idea of what you were working ( your task list is your primary source of context between context compressions , so it needs to be we managed )
- show the tasks list after completing a task
- ensure you update your context before working on the next task


( continue working ) ( get to work ) 
