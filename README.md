# PoC for Temporal

## Quick run

- Start temporal server
```bash
docker compose down && docker compose up
```
- In separate terminal, start worker
```bash
go run ./paydetails
```
- In separate terminal, start workflow
```bash
alias temporal_docker="docker exec temporal-admin-tools temporal"
temporal_docker workflow start \
 --task-queue backgroundcheck-boilerplate-task-queue-local \
 --type BackgroundCheck \
 --input '"555-55-5555"'
```

In the terminal with worker, you should see something happening. You should also be able to see your workflows at http://localhost:8080/namespaces/default/workflows.
