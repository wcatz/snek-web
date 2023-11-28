      // Keep track of the previous slot number
      let prevSlotNumber;
      // Event handler for incoming messages
      function handleWebSocketMessage(event) {
        // Parse the JSON message
        const blockEvent = JSON.parse(event.data);

        // Create a new div element for each message block
        const newDiv = document.createElement("div");
        newDiv.classList.add("message-block");
        newDiv.id = "message-" + new Date().getTime(); // Unique ID based on timestamp

        // Calculate time difference between consecutive messages based on slot number
        let timeDifference = "";
        if (prevSlotNumber !== undefined) {
          const slotsDiff = blockEvent.context.slotNumber - prevSlotNumber;

          if (slotsDiff < 60) {
            // If less than 60 seconds, display only seconds
            timeDifference = `${slotsDiff} Ssecondsss`;
          } else {
            // If more than 60 seconds, convert to minutes and seconds
            const minutes = Math.floor(slotsDiff / 60);
            const remainingSeconds = slotsDiff % 60;
            timeDifference = `${minutes} Minutesss ${remainingSeconds} Ssecondssss`;
          }
        }
        // Convert block size to kilobytes
        const blockBodySize = blockEvent.payload.blockBodySize;
        const blockBodySizeKB = blockBodySize / 1024;

        // Full block size in kilobytes
        const fullBlockSize = 90112;

        // Calculate percentage full based on kilobytes
        const percentageFull = (blockBodySizeKB / (fullBlockSize / 1024)) * 100;

        // Populate the new div with content
        newDiv.innerHTML = `
<div class="zoom-in mx-auto max-w-5xl px-4 sm:px-6 lg:px-8 bg-black p-6 rounded-lg shadow-lg">
  <pre class="whitespace-pre-line text-green-600 mx-auto">
    <span class="text-blue-400">Type</span><span class="text-white">:</span> ${
      blockEvent.type
    }
    <span class="text-blue-400">Timestamp</span><span class="text-white">:</span> ${
      blockEvent.timestamp
    }
    <span class="text-blue-400">Block Number</span><span class="text-white">:</span> ${
      blockEvent.context.blockNumber
    }
    <span class="text-blue-400">Slot Number</span><span class="text-white">:</span> ${
      blockEvent.context.slotNumber
    }
    <span class="text-blue-400">Block Size</span><span class="text-white">:</span> ${blockBodySizeKB.toFixed(
      2
    )} KB, ${percentageFull.toFixed(2)}% full
    <span class="text-blue-400">Pool</span><span class="text-white">:</span> <span class="whitespace-pre-line text-wrap" id="issuerVkey">${
      blockEvent.payload.issuerVkey
    }</span>
    <span class="text-blue-400">Block Hash</span><span class="text-white">:</span> <span class="whitespace-pre-line text-wrap" id="blockHash">${
      blockEvent.payload.blockHash
    }</span>
    <span class="text-blue-400">Transaction Count</span><span class="text-white">:</span> ${
      blockEvent.payload.transactionCount
    }
  </pre>
  <div class="flex justify-center gap-1 mb-2">
    <button onclick="copyToClipboard('${
      newDiv.id
    }', 'issuerVkey')" class="copy-button mt-2 border border-green-600 hover:border-green-700 text-blue-400 hover:text-blue-500 py-1 px-2 rounded focus:outline-none focus:border-green-700 active:border-green-900 active:text-blue-600">
      Copy Pool
    </button>
    <button onclick="copyToClipboard('${
      newDiv.id
    }', 'blockHash')" class="copy-button mt-2 border border-green-600 hover:border-green-700 text-blue-400 hover:text-blue-500 py-1 px-2 rounded focus:outline-none focus:border-green-700 active:border-green-900 active:text-blue-600">
      Copy Hash
    </button>
  </div>
</div>
<div class="text-center py-1 text-white">${timeDifference}</div>
`;

        // Update the previous slot number for the next message
        prevSlotNumber = blockEvent.context.slotNumber;

        document
          .getElementById("messages-container")
          .insertBefore(
            newDiv,
            document.getElementById("messages-container").firstChild
          );
      }

      // Create a WebSocket connection
      const socket = new WebSocket("ws://localhost:8080/ws");

      // Event handler for when the WebSocket connection is established
      socket.addEventListener("open", (event) => {
        console.log("WebSocket connection established:", event);
      });

      // Event handler for WebSocket errors
      socket.addEventListener("error", (event) => {
        console.error("WebSocket error:", event);
      });

      // Event handler for WebSocket closure
      socket.addEventListener("close", (event) => {
        console.log("WebSocket connection closed:", event);
      });

      // Event handler for incoming messages
      socket.addEventListener("message", handleWebSocketMessage);

      socket.addEventListener("message", function (event) {
        if (event.data === "refresh") {
          window.location.reload();
          return;
        }
      });

      function copyToClipboard(divId, property) {
        const element = document.getElementById(divId);
        const propertySpan = element.querySelector(`span#${property}`);
        const propertyValue = propertySpan.innerText;

        const tempInput = document.createElement("textarea");
        tempInput.value = propertyValue;
        document.body.appendChild(tempInput);
        tempInput.select();
        document.execCommand("copy");
        document.body.removeChild(tempInput);

        // Highlight copy
        propertySpan.classList.add(
          "text-green-300",
          "transition-color",
          "duration-500"
        );
        setTimeout(() => {
          propertySpan.classList.remove("text-green-300");
        }, 500);
      }
      function updateNodeAddress() {
        var newNodeAddress = document.getElementById("nodeAddressInput").value;

        fetch("/updateNodeAddress", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(newNodeAddress),
        })
          .then((response) => response.json())
          .then((data) => {
            console.log("Node address updated:", data);
          })
          .catch((error) => {
            console.error("Error updating node address:", error);
          });
      }
      // Fetch the current node address on page load
      document.addEventListener("DOMContentLoaded", function () {
        fetch("/getNodeAddress")
          .then((response) => response.json())
          .then((data) => {
            if (data.nodeAddress) {
              document.getElementById("nodeAddressInput").value =
                data.nodeAddress;
            }
          })
          .catch((error) =>
            console.error("Error fetching node address:", error)
          );
      });