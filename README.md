# DS-Assignment-2

This project contains the implementation of the following mutual exclusion protocols:
1. Fair Ring Protocol
2. Lamport's shared priority queue with Ricart-Agrawala optimization.
3. Voting Protocol with deadlock avoidance.

## How to run the program:

1. Change into the directory of the protocol you want to run.
2. Open 10 powershell windows in this directory and run the program using the command provided in the next step in each of the windows.
3. To run the program, find the executable build file and run it using one of the appropriate the commands for the chosen protocol:

For Fair Ring Protocol:
```powershell
./fair-ring.exe
```
For Lamport's shared priority queue with Ricart-Agrawala optimization:
```powershell
./lamport-shared-queue.exe
```

For Voting Protocol with deadlock avoidance:
```powershell
./voting-protocol.exe
```

4. The first powershell window(bootstrap node) will ask for the number of requests to be made. Enter the number of requests and press enter.
5. After making sure that all the powershell windows are successfully running the RPC servers for each node, enter y in the bootstrap node to start the requests.

The protocol should begin execution and description regarding the different events will be printed on the console. Once the first critical section of all the nodes requesting is completed, the program will terminate and will display the time taken until the last critical section is executed.

Each output has the following format:
```
[NODE-ID] [EVENT]
```
![image](https://github.com/user-attachments/assets/7b8331c8-17b3-4963-a0c7-65f49e4f3704)

![image](https://github.com/user-attachments/assets/98701585-e41b-4a7b-aa22-eb714495324d)

## Analysis of the protocols and their performance:

The following analysis is done based on the number of requests made by the nodes and the time taken to complete these requests.

However, before we dive into the analysis, there are a few considerations to keep in mind:
1. For the purposes of simulating an actual critical section, a dummy critical section function was created which takes 2 seconds to execute for all protocols.

![image](https://github.com/user-attachments/assets/129ffe1a-b219-40b8-9439-fe6c1c7836bf)

2. There is a time delay of 1 second between each message received by the nodes to slow down the print statements and make it easier to understand the flow of the program.

![image](https://github.com/user-attachments/assets/c46b8ec5-1a1e-4b46-86c8-ff1cf6052f33)

Below is the result of the analysis of the protocols:

Table data: (Time taken in seconds)
| Number of requests | Fair Ring Protocol | Lamports shared priority queue + ricart-agrawala optimization | Voting Protocol |
|---------------------|--------------------|-------------------------------------------------------------|-----------------|
| 1                   | 23.38s             | 21.99s                                                      | 4.37s            |
| 2                   | 35.46s             | 23.94s                                                      | 12.4s            |
| 3                   | 47.21s             | 28.15s                                                      | 28.4s            |
| 4                   | 59.52s             | 33.09s                                                      | 39.16s           |
| 5                   | 71.69s             | 39.02s                                                      | 36.75s           |
| 6                   | 84.4s              | 46.01s                                                      | 54.47s           |
| 7                   | 95.33s             | 53.57s                                                      | 68.15s           |
| 8                   | 107.34s            | 62.74s                                                      | 78.48s           |
| 9                   | 119.36s            | 72.83s                                                      | 86.95s           |
| 10                  | 131.5s             | 84.45s                                                      | 91.18s           |

Graph data:

![image](https://github.com/user-attachments/assets/f320f686-3b39-4513-bc6e-d1fd19f42aa4)

From the graph, we can decipher that the Fair Ring Protocol is the slowest amongst the three protocols. But when it comes to Lamport's shared priority queue, it is the most consistent protocol in terms of time taken to complete the requests as the number of requests increases. The Voting Protocol with deadlock avoidance is the fastest when the number of requests is small, but when the number of requests increases, the time taken to complete the requests increases much faster than the Lamport's shared priority queue.
