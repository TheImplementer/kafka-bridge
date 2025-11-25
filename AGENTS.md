# Repository Guidelines

## Project Structure & Module Organization
The bridge is split into JVM modules: `bridge-runtime/` (common APIs), `http-bridge/` (REST + WebSocket adapters), `kafka-clients/` (producer/consumer wrappers), and `integration-tests/`. Source lives under each module's `src/main/java` and `src/main/resources`; unit tests live in `src/test/java`, while long-running suites sit in `src/it/java`. Sample payloads, Postman collections, and compose files reside in `examples/` and `infra/`. Keep assets (schemas, JSON fixtures) near the feature they serve to simplify reviews.

## Build, Test, and Development Commands
- `mvn clean install` – compiles all modules, runs unit tests, and produces runnable JARs under `*/target/`.
- `mvn -pl http-bridge spring-boot:run` – starts the HTTP bridge with hot reload against your local Kafka cluster.
- `mvn verify -Pintegration-tests` – executes Docker-based integration suites defined in `integration-tests/pom.xml` (requires Docker Desktop).
- `docker compose -f infra/docker-compose.yml up kafka bridge` – spins Kafka, ZooKeeper, and a dev bridge for manual testing.

## Coding Style & Naming Conventions
Use Java 17, 4-space indentation, and favor explicit types. Classes/interfaces follow `UpperCamelCase`, methods `lowerCamelCase`, constants `UPPER_SNAKE_CASE`. Topics, consumer groups, and configs in samples should stay kebab-case (e.g., `orders-stream`). Run `mvn spotless:apply` and `mvn checkstyle:check` before opening a PR; never commit generated code without adding a `.gitattributes` rule.

## Testing Guidelines
Unit tests belong beside the code they exercise, using JUnit 5 and AssertJ. Mock Kafka interactions with `KafkaTestUtils`; reserve real broker calls for integration tests tagged `@Tag("it")`. Name tests `ClassNameTest` and integration suites `ClassNameIT`. Target ≥80% line coverage for new modules and document manual verification steps in `docs/testing.md` when automation is not possible.

## Commit & Pull Request Guidelines
Adopt Conventional Commits (`feat:`, `fix:`, `chore:`). Keep messages under 72 characters in the subject and include context or Jira links in the body. Each PR should describe the bridge scenario solved, list test commands executed, reference an issue ID, and attach logs or screenshots if behavior changes. Request reviews from ownership code owners listed in `CODEOWNERS`, and wait for green CI plus at least one approval before merging.

## Security & Configuration Tips
Store Kafka credentials in `.env.local` (ignored by git) and load them through Spring's config placeholders. Never check in cloud secrets; provide sanitized examples via `config/example-bridge.properties`. Rotate SASL credentials in examples quarterly and reference the shared Vault path when updating CI secrets.
