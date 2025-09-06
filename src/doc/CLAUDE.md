


- IMPORTANT AXIOM RULE TO FOLLOW : outside of unit tests , consider mocks and stubs the worst architectural anti pattern ( stop metioning the fact that you are using real implmentations , this is explicit in the implementation , keep the naming professional )

- IMPORTANT AXIOM RULE TO FOLLOW : we cannot create new files outside of our current folders and files structure ( this is deliberate ) . you may alter implmentation details . you may ask to delete/create files if you have a proper reason to do so 

- IMPORTANT AXIOM RULE TO FOLLOW : do not edit any markdown files without permission . ensure you stop and ask for permission and provide your reasoning

- IMPORTANT AXIOM RULE TO FOLLOW : Our database schemas and endpoints and clients should exactly match that in our TABLES-[name].md markdown files ( the endpoints are in the API-GATEWAYS.md file ) ( these are located in the doc directory )

- IMPORTANT AXIOM RULE TO FOLLOW : consider stubs the worst anti-pattern ( we should use propper infrastructure and fix root issues as they arise ) ( we should not fall back to stubs in integration tests ) 

- IMPORTANT AXIOM RULE TO FOLLOW : only run integration tests when the entire development environment is up ( only the deployer implements integration tests ) 

- note : stop using the term 'medical-grade' ( this is implicit and there is no need to mention it in the naming system )

- note : do not work with github actions for now

- note : currently we do not own a website domain in cloudflare 



[deployment secrets]

- you can use the authenticated azure cli for deployment with pulumi

CLOUDFLARE_ACCOUNT_ID="a71f1b4cc92fea4a215a8468b3390350"
CLOUDFLARE_API_TOKEN="WjlWxu5Lz-DwSPHXB8NlNcB4iMjd_uvR7Q06-bI-6HY.5YIkJX1L_OgIJqwrkasR"
CLOUDFLARE_ZONE_ID="2bdbd19afb3154b1768c0a80bc65b7d9"
CLOUDFLARE_ZONE_NAME="axiomcloud.dev"

HASHICORP_CLIENT_ID="lfNhxsGjHnhffHi9Ze5qJg39LjzRdwxX"
HASHICORP_CLIENT_SECRET="i2KHbAbH9BwKnm2-olJ4-R5pO5rrjMK_x4WOH3rG2U-jURNW8CyWPaaxrg1LTIZT"

UPSTASH_API_KEY="bcf68703-8e1a-4611-a21d-aebd2857f437"

CLOUD_AMQP="46c20467-552b-4645-9189-47f69bc78df8"

GRAFANA_CLOUD_ACCESS_POLICY_TOKEN="glc_eyJvIjoiMTA3Nzk3NSIsIm4iOiJwdWx1bWktcG9saWN5LXB1bHVtaS1wb2xpY3ktdG9rZW4iLCJrIjoiNUNCbHY5bjZaWHFIQko3MzUxb0EwUDcxIiwibSI6eyJyIjoidXMifX0="



[ important project related rules to follow ]

# development workflow

- test-driven development ( red phase , green phase , refactor phase ) ( tests drive and validate the design of our architecture ) ( creating new methods from refactoring should not require us to write new tests , this violates the contract-first testing principle ) ( you are allowed to modify the project and tests implementations as you see fit , since project and/or tests abstractions and/or implementations sometimes need to be updated ) ( when planning a new TDD cycle , provide a list of all the files you intend to edit and what you intend to do in each phase ) ( you may alter existing tests in new tdd cycles if there are good reasons to do so )

- use pulumi for local development environment ( using podman instead of docker ) 



# architecture layers

- the highest layer is the frontend
- the middle layer is the backend services 
- the lowest layer is the deployment environments layer



# architecture patterns

- cohesion over coupling
- seperation of concerns

- best practices in our stack

- http server patterns 
- handler , service , repository pattern, dapr-centric
- dependency inversion ( this does not mean interfaces everywhere )
- synchronous and asynchronous patterns were appropriate

- warnings as errors
- singletons can only depend on singletons

- be causious of implementating factories ( they create more complexity than the issues they usually solve ) 

- we should discover abstractions based on duplicated implementations instead of trying to predict them ( this makes it less likely that we will implment complexity that we did not anticipate ) ( premature abstractions are usually not helpful ) 

- the result pattern is an anti pattern ( go has built in error handling ) 
- dont repeat yourself ( DRY ) is not good and can create more problems than the problems it solves ( it is okay to repeat code if it reduces complexity ) ( simplicity over complexity ) ( only consider abstracting implmentations when there is clear excessive duplication )



# website

- public api gateway for dynamic data

- astro
- vue
- pinia

- tailwind
- shadcn-vue

- nodejs runtime
- pnpm ( do not use npm nor bun nor dino nor npx ) 

- vite
- vitest
- headles playwright

- do not use react
- do not do UI design testing

## testing

- properly-based testing
- any tests that rely on clients that connect to the backend should use mocks ( all tests are isolated to the website without the backend . we perform testing with the real backend in cicd instead ) 

- do not change the UI of our website unless explicitly asked to 


# backend

- golang

- idiomatic go patterns

- dapr resources apis
- slog for logging

## api gateways

- handle cross cutting concerns

- reverse proxy
- rate limiting
- security headers
- cors policies

- security ( we must have fallback policies that get evaluated if no other policy is specifieid )

### public

- anonymous usage
- ip-based rate limiting ( 1000 req/min )
- public website origins
- azure cosmosdb backing store
- standard observability
- standard security

### admin

- role-based access control
- user-based rate limiting ( 100 req/min )
- medical-grade audit loggging
- medical-grade security

## api domains

- domain shared kernels for public and admin apis
- vertical slice public and admin apis

## observability

- logging ( use structured logging , not concatenation ) ( each log should have key bits of information ( user ID , Correrlation ID , Request URL , APP Version , and so forth ) ( logs should be developer focused ) ( log levels : debug , information , waarning , error , critical ) ( not having 100% log delivery is okay )

- admin gateway audits ( for medical-grade compliance ) ( losing any data is unacceptable ) ( store in grafana cloud loki )

## testing

- properly-based testing
- unit tests must use mock for dependencies to craete isolation



# cicd

- golang-migrate for migrations
- sql migration files 

- feature flags ( we will work on this later ) 

- infrastructure as code patterns ( stack per environment ( shared , dev , staging , prod ) ) ( component-first architecture ) ( no hard-coded secrets ) ( least privilage IAM ) ( consistent naming conversion ) ( pulumi testing framework for unit tests , property-based tests , integration tests )  ( Automation API for Programmatic Infrastructure Management for CICD workflows ) ( integration testing framework "github.com/pulumi/pulumi/pkg/v2/testing/integration" ) ( reproducibility over ephimeral solutions ) 

## database migrations

- automated development and testing migrations
- manual production migrations## dapr stateless services

- note : azure container apps manages dapr ( we have to create it manually in development environment for a similar environment ) 

- orchestrator : dapr control plan container for production and local development 

- apis services : api service and dapr sidecar containers for production and local development 
- gateway services : gateway service and dapr sidecar containers for production and local development

- identity provider : authentik container for production and local development 
- telemetry collection : grafana agent and dapr sidecar containers for production and local development

- infrastructure configuration : yaml configuraation files for production and local development

## dapr sidecar middleware configurations

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

- pub/sub and authenticated sessions : upstash redis hosted for production ( redis container for local development ) 
- secret store : hashicorp vault cloud hosted for production ( hashicorp vault and vault data containers for local development )

- relational database ( includes configuration store and state store and identity storage and services storage ) : azure manged postgre hosted for production ( postgre container for local development ) 
- file storage : azure blob storage hosted for production ( azurite blob storage emulator and azurite-data containers for local development https://github.com/Azure/Azurite )

## telemetry observability 

- objeservability : grafana , mimir , loki , tempo , pyroscope in container with their respective data volumes containers ( grafana-data , mimir-data , loki-data , tempo-data , pyroscope-data ) for local development ( grafana cloud for production ) 

## website content delivery

- website hosting : cloudflare pages
- content delivery network : cloudflare cdn

## persistent storage migrations

- golang-migate in pulumi

- development :
    approach : Aggressive - always migrate to latest
    rollback : Easy - can destroy and recreate
    safety_checks : Minimal
    automation : Full automation
    
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

- pulumi golang

- state storage : azure blob storage ( pulumi state storage container ) ( manually created using the azure cli ) 

- github secrets : ( azure authentication , pulumi state storage , grafana cloud authentication , hashicorp vault cloud authentication ) ( manually created using the azure cli , cloudflare cli , grafana cli , hashicorp vault cli , github cli ) 

- github container registry

## testing

- unit tests must use mock for dependencies to craete isolation
- integration tests must use real dependencies ( not mocks ) ( health validation ) ( contract tests ) ( environemnt health ) 

- end to end tests ( we will implment these some other time , not now ) 
- performance tests ( we will implment these some other time , not now )
- chaos engineering tests ( we will implment these some other time , not now )



# testing ( website , backend , cicd )

- arrange , act , assert

- contract-first testing ( testing interfaces/contracts rather than implementation details ) ( focused on preconditions/dependencies and postconditions/state-change )
- environment-aware testing ( this applies to all non-unit tests ) ( in-memory , development , staging , production ) ( critical : we will not implement this yet . we will do it some other time ) 

- testing results should be reprodusable ( temporary fixes need to followed up with reprodusable implmentation )
- all tests must have timeouts ( they should fail fast if something is wrong ) ( 5 seconds for unit tests ) ( 15 secnds for integration ) ( 30 seconds for end to end tests )

- do not use curl commmands or cli tools for testing ( test through our testing framework )



# version control

- trunk based development for version control

- commit message template ( will implement this some other time ) 

- github version control



# api versioning

- 



[ important general rules to follow ]

- ultrathink and deep dive and be comprehensive and be professional

- this is nixos and the admin password is 'unsecure'

- fix issues as they arise 

- be causious of technical debt
- be causious of overengineering

- do not create documentation
- do not preserve legacy implementations
- do not implement experimental architectures not part of industry
- do not create script files for projects ( this is an anti-pattern )
- do not create simple ephimeral validation implementations in /temp/ directories to avoid disorder in source files

- critical axiom rule : do not stage and commit and push unless I explicitly ask you to

# Task Management Context Guidelines

- task descriptions must include WHY ( business reason , compliance requirement , architectural decision , so forth )
- task descriptions must include SCOPE ( which APIs , which components , which environments , so forth )
- task descriptions must include DEPENDENCIES ( what must complete first , integration points , so forth )
- task descriptions must include CONTEXT ( gateway architecture , environment specifics , so forth )

- note : when a plan gets approved with the word 'approved' , add all the tasks to your tasks list . 
- note : in the event context compression happens in the middle of a task , your task list is your primary source of context between context compressions , so it needs to be well managed 

( continue working ) ( get to work ) 
