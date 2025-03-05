# **Fraud Prevention System**  

This project implements a **fraud prevention system** using a microservices architecture in **Golang**.  

- **Compliance Service:** Manages stolen card reports and maintains a list of blocked credit cards.  
- **Payment Service:** Processes transactions and checks whether a card is blocked before approving payments.  

Both services communicate via **REST API** and are fully containerized with **Docker**.  

## **How can I test it?**  
### **1. Run with Docker Compose**
```
docker-compose up --build
```
This will start compliance-service (port 8080) and payment-service (port 8081).

### **2. Report a Stolen Card**  
1. Open your browser and visit: **[`http://localhost:8080/report`](http://localhost:8080/report)**  
2. Enter the following credentials:  
   - **Username:** `john_doe`  
   - **Secret Code:** `hashed_secret_123`  
3. Submit the report.  

### **3. Attempt a Payment (Blocked if Stolen)**  
Run the following **CURL** command to simulate a payment request:  

```bash
curl --location 'http://localhost:8081/process_payment' \
--header 'Content-Type: application/json' \
--data '{
    "user_id": 1,
    "card_id": 2,
    "amount": 100.50
}'
```
