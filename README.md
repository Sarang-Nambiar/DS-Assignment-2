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

<!-- show sample output here -->

## Analysis of the protocols and their performance:

The following analysis is done based on the number of requests made by the nodes and the time taken to complete these requests.

However, before we dive into the analysis, there are a few considerations to keep in mind:
1. For the purposes of simulating an actual critical section, a dummy critical section function was created which takes 2 seconds to execute for all protocols.
<!-- insert the image of critical section here -->
2. There is a time delay of 1 second between each message received by the nodes to slow down the print statements and make it easier to understand the flow of the program.
<!-- insert image of the time delay snippet -->

Below is the result of the analysis of the protocols:

Table data: (Time taken in seconds)
| Number of requests | Fair Ring Protocol | Lamports shared priority queue + ricart-agrawala optimization | Voting Protocol |
|---------------------|--------------------|-------------------------------------------------------------|-----------------|
| 1                   | 23.38             | 21.99                                                      | 4.37            |
| 2                   | 35.46             | 23.94                                                      | 12.4            |
| 3                   | 47.21             | 28.15                                                      | 28.4            |
| 4                   | 59.52             | 33.09                                                      | 39.16           |
| 5                   | 71.69             | 39.02                                                      | 36.75           |
| 6                   | 84.4              | 46.01                                                      | 54.47           |
| 7                   | 95.33             | 53.57                                                      | 68.15           |
| 8                   | 107.34            | 62.74                                                      | 78.48           |
| 9                   | 119.36            | 72.83                                                      | 86.95           |
| 10                  | 131.5             | 84.45                                                      | 91.18           |

Graph data:
<!-- insert graph here -->

From the graph, we can decipher that the Fair Ring Protocol is the slowest amongst the three protocols. But when it comes to Lamport's shared priority queue, it is the most consistent protocol in terms of time taken to complete the requests as the number of requests increases. The Voting Protocol with deadlock avoidance is the fastest when the number of requests is small, but when the number of requests increases, the time taken to complete the requests increases much faster than the Lamport's shared priority queue.



<!-- insert the image of the table here with a graph below it -->