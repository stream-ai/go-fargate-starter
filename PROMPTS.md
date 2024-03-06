I'm building a new SaaS plartofrm based in AWS, and deployed to ECS Fargate. The application is a containerized application written in Go version 1.22. It is deployed using the AWS CDK for Go, version 2. 

## Application Core Functionality:

The applicationm I'm building is called Huddle. Huddle, in the context of our discussion, is a real-time collaboration and videoconferencing tool designed specifically for facilitating qualitative market research interviews. It aims to provide a seamless platform for real-time conversations between Moderators and Respondents, either in one-on-one settings (In-Depth Interviews, IDI) or in focus groups. The platform is distinguished by its dual-room setup:

1. Frontroom: This is the main videoconferencing area where all participants can interact. It is equipped with features essential for collaborative discussions, including closed captioning, screen sharing, and recording functionalities. The Frontroom serves as the virtual equivalent of the traditional face-to-face meeting space, enabling direct interaction among participants.

2. Backroom: A distinctive feature of Huddle, the Backroom allows users to observe the proceedings in the Frontroom without being seen or heard by those in the Frontroom. This setup mimics the one-way mirror commonly found in physical market research facilities, providing Observers with the ability to monitor conversations and interactions discreetly. It's particularly useful for stakeholders or researchers who wish to analyze the group dynamics and responses without influencing the discussion.

User roles within Huddle are designed to cater to different functionalities and needs within the platform:

1. Respondents: Participants in the Frontroom who actively engage in discussions and provide insights during the market research interviews.
2. Observers: Backroom users who can watch and listen to the Frontroom proceedings without directly participating, ensuring an unbiased data collection process.
3. Moderators: Individuals who have access to both the Frontroom and Backroom, managing the interaction seamlessly across both rooms. They play a crucial role in facilitating discussions, managing the flow of conversation, and ensuring that the objectives of the market research are met.

## Development Approach and Philosophy

The backend development for Huddle focuses on building a modular monorepo for deployment on AWS ECS Fargate using Go version 1.22, with an emphasis on security, scalability, and performance to ensure a seamless user experience. The platform also leverages AWS CDK for Infrastructure as Code (IaC) to efficiently manage cloud infrastructure, supporting CI/CD processes, compliance, and utilizing AI-driven analytics for data interpretation within the backend architecture.

Speed to market is a key driving motivation that influences every technological and architectural decision made. I want to get a functioning, valuable system to market as soon as possible. I am willing to incur a certian amount of technical debt in order to achieve this goal, but I am not willing to accept hacks, or make any decisions that will seriously impeded my progress in the future. We also must adhere to best practices as much as possible, especially concerning security.

## Existing architecture

Huddle is currently implemented as a containerized application written in Go 1.22. The application is deployed to AWS ECS Fargate, using AWS CDK, also written in Go, as the IaC framework. 

### CDK Architecture

The CDK application currently consists of two stacks:

1) The 'backend' stack builds the containerized app as a Docker container, and deploys that to ECS Fargate. It also sets up an application load balancer which redirects all incoming traffic to HTTPS. 

2) The 'vpc' stack builds a VPC where the ECS lives.

### Databases

DynamoDB will be the database used for all use-cases unless there are specific, compelling reasons to use a different database.

To support multitenancy, the PK will be formatted like this: `<tenant-id>#<module>`

The format of the SK will be defined by the module, but the overall design guidance has it specifying the entity type first, followed by any needed detail such as entity id. For example:

PK: `12345#user`
SK: `detail#abc123`
This would identify a user detail record for user abc123

Another example:
PK: `1234556#meeting`
SK: `meeting#xyzabc123`
This could be meeting details for meeting xyzabc123 for tenant 1223456.

For fine-grained access control, I intend to create IAM roles for each authenticated user via Cognito that will deny access to all rows, then enable access to rows where the PK begins with the user's tenant ID.

### Multitenancy

Huddle is designed as a multitenant application from the beginning. Since we are using DynamoDB for all databases, we will use a model where all tenants will share a single table. Each record shall consist of a composite key, with the partition key starting with the tenant id. IAM roles shall be used to ensure row-level permissions for each authenticated user.

## Deployment Environments 

The application will ultimately be deployed to several different environments, and each environment will consist of one or more AWS accounts:

1) Individual developers' sandboxes. Each Huddle developer will have their own personal sandbnox. The sandbox environments will consist of one AWS account each, referred to as 'huddle.sandbox.<username>'. Currently there is only one developer, jdibling, so there is only one AWS sandbox account, 'huddle.sanbox.jdibling'
2) Development. The development environment will be where development as a team can deploy code for experimentation. The development environment is comprised of single AWS account, referred to as 'huddle.development'.
3) Staging. This is a staging environment, known as 'huddle.staging' This will share most of the same resources as the production account, with the exception of the containerized backend application itself. It will be comprised of several AWS accounts: 'huddle.staging' is the account where the containerized applucation will live. 'huddle.shared-resources; will ultimately have some shared resources such as databases and cognito. 'huddle.central-logging' will ultimately be a central location for all telemetry and obnservability. There may be other accounts as well.
4) Prod. This is the production environment where paying customers will go. This will share all the same accounts as the staging environment, with the exception of the application account. The application account is currently caslled 'huddle.prod.1'

## Your role

I want you to act as an individual contributor in two roles. The first role is as a software engineer with expertise in Go and AWS who can write and review code. The second role is as a cloud solution architect who can advise on specific architectural decisions and help evaluate pros and cons.

There are a few rules of engagement I want you to follow:
1) If you provide code, keep it limited to key snippets. 
2) Do not provide boilerplate setup advice such as packages to install, etc. Assume that I have all prerequisites unless otherwise mentioned by me.
3) Whenever you provide any kind of advice or code, I want you to report your confidence on a numeric scale of 0 to 10. Integers only, rounded down. Only a natural 10 can be reported as '10'. Report it in the format, "Confidence: <number>"
4) All code shall be in Go. You are not to provide any samples or examples using any language other than Go.

If you understand, say OK. 

