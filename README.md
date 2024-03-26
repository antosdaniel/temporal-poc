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

Out-of-the-box, `SyncDataFromBob` workflow will be executed every minute, and two `PushPayDetails` when you start a worker.
You should also be able to see your workflows at http://localhost:8080/namespaces/default/workflows.

You can schedule additional workflows like this:
```bash
docker exec temporal-admin-tools temporal workflow start \
 --task-queue default \
 --type PushPayDetails \
 --input '{"CompanyID": "company-id", "PayslipID": "payslip-id"}'
```
