function copyToClipboard(element, property) {
  const propertySpan = element;
  const textToCopy = propertySpan.innerText;

  // Use the Clipboard API if available
  if (navigator.clipboard) {
    navigator.clipboard.writeText(textToCopy).then(() => {
      // Highlight copy
      propertySpan.classList.add(
        "text-white",
        "transition-color",
        "duration-300"
      );
      setTimeout(() => {
        propertySpan.classList.remove("text-white");
      }, 500);
    }).catch(err => {
      console.error('Unable to copy to clipboard', err);
    });
  } else {
    // Fallback for browsers that do not support the Clipboard API
    const tempInput = document.createElement("textarea");
    tempInput.value = textToCopy;
    document.body.appendChild(tempInput);
    tempInput.select();
    document.execCommand("copy");
    document.body.removeChild(tempInput);

    // Highlight copy
    propertySpan.classList.add(
      "text-white",
      "transition-color",
      "duration-300"
    );
    setTimeout(() => {
      propertySpan.classList.remove("text-white");
    }, 500);
  }
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
    .catch((error) => console.error("Error fetching node address:", error));
});

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

function updateEventType(selectedEventType) {
  if (selectedEventType) {
    fetch("/updateEventType", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(selectedEventType),
    })
      .then((response) => response.text())
      .then((data) => {
        console.log("Event type updated:", data);
      })
      .catch((error) => {
        console.error("Error updating event type:", error);
      });
  } else {
    console.error("Selected event type is null or undefined");
  }
}

document.addEventListener("DOMContentLoaded", function () {
  const toggleButton = document.getElementById("toggleButton");
  const closeButton = document.getElementById("closeButton");
  const offcanvasMenu = document.querySelector(".offcanvas");
  const toggleMenu = function () {
    offcanvasMenu.classList.toggle("transform-none");
  };

  // Toggle menu when clicking the toggle button
  toggleButton.addEventListener("click", toggleMenu);

  // Close menu when clicking the close button
  closeButton.addEventListener("click", toggleMenu);

  // Close menu when clicking outside the menu
  document.addEventListener("mousedown", function (event) {
    if (
      !offcanvasMenu.contains(event.target) &&
      event.target !== toggleButton
    ) {
      offcanvasMenu.classList.remove("transform-none");
    }
  });

  // Close menu when pressing the Escape key
  document.addEventListener("keydown", function (event) {
    if (event.key === "Escape") {
      offcanvasMenu.classList.remove("transform-none");
    }
  });

  // Close menu when clicking on the page (outside the menu or toggle button)
  document.addEventListener("focusin", function (event) {
    if (
      !offcanvasMenu.contains(event.target) &&
      event.target !== toggleButton
    ) {
      offcanvasMenu.classList.remove("transform-none");
    }
  });
});
// When the user scrolls down 2 screen heights, show the button
window.onscroll = function () {
  const scrollThreshold = window.innerHeight * 2;
  const toTopButton = document.getElementById("toTop");

  if (
    document.body.scrollTop > scrollThreshold ||
    document.documentElement.scrollTop > scrollThreshold
  ) {
    toTopButton.classList.remove("hidden");
  } else {
    toTopButton.classList.add("hidden");
  }
};
// When the user clicks on the button, scroll to the top of the document
function goToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

// Function to calculate time difference based on slot number
function calculateTimeDifference(currentEvent) {
  let timeDifference = "";

  // Check if previous slot number is defined
  if (prevSlotNumber !== undefined) {
    const currentSlotNumber =
      currentEvent.type === "chainsync.block"
        ? currentEvent.context.slotNumber
        : currentEvent.type === "chainsync.rollback"
          ? currentEvent.payload.slotNumber
          : currentEvent.type === "chainsync.transaction"
            ? currentEvent.context.slotNumber
            : 0; // Assuming slot number for transaction type is available in context

    // Only calculate time difference if transitioning from a block to another
    if (
      (currentEvent.type === "chainsync.block" ||
        currentEvent.type === "chainsync.transaction") &&
      prevSlotNumber !== currentSlotNumber
    ) {
      const slotsDiff = Math.abs(currentSlotNumber - prevSlotNumber);
      timeDifference = formatTimeDifference(slotsDiff);
    }
  }

  // Update the previous slot number for the next message
  prevSlotNumber =
    currentEvent.type === "chainsync.block"
      ? currentEvent.context.slotNumber
      : currentEvent.type === "chainsync.rollback"
        ? currentEvent.payload.slotNumber
        : currentEvent.type === "chainsync.transaction"
          ? currentEvent.context.slotNumber
          : prevSlotNumber; // Keep the previous value for unknown event types

  return timeDifference;
}

// Function to format time difference in a human-readable format
function formatTimeDifference(slotsDiff) {
  if (slotsDiff < 60) {
    return `${slotsDiff} seconds`;
  } else {
    const minutes = Math.floor(slotsDiff / 60);
    const remainingSeconds = slotsDiff % 60;
    return `${minutes} minutes ${remainingSeconds} seconds`;
  }
}

// Function to handle left arrow click (Go Back)
function goBack() {
  if (inExtraDataView) {
    // Reset the view state to "main page" view only if inExtraDataView is true
    inExtraDataView = false;

    // Clear the screen
    clearScreen();

    // Repaint the
    Array.from(eventsMap.values()).forEach(({ eventData, newDiv }) => {
      displayEvent(eventData, newDiv);
      insertMessageDiv(newDiv);
    });
  }
  removeGoBackButton(); // Remove the "Go Back" button
}


// Function to remove the "Go Back" button
function removeGoBackButton() {
  const goBackButton = document.getElementById("goBackButton");
  if (goBackButton) {
    goBackButton.parentNode.removeChild(goBackButton);
  }
}

// Function to handle right arrow click
function handleRightArrowClick(transactionEvent) {
  // Store the current scroll position
  const currentScrollPosition = window.scrollY;
  // Set the view state to "extra data" view
  inExtraDataView = true;
  // Clear the screen
  clearScreen();
  // Display extra data
  displayExtraData(transactionEvent);
  // Scroll to the top of the page
  window.scrollTo({ top: 0, behavior: "smooth" });
}

// Function to display BlockEvent
function displayBlockEvent(blockEvent, newDiv) {
  // Convert block size to kilobytes
  const blockBodySize = blockEvent.payload.blockBodySize;
  const blockBodySizeKB = blockBodySize / 1024;
  // Full block size in kilobytes
  const fullBlockSize = 90112;
  // Calculate percentage full based on kilobytes
  const percentageFull = (blockBodySizeKB / (fullBlockSize / 1024)) * 100;
  // Calculate time difference
  const timeDifference = calculateTimeDifference(blockEvent);
  newDiv.innerHTML = `
    <div class="zoom-in mx-auto max-w-5xl px-4 sm:px-6 lg:px-8 bg-black p-6 rounded-lg shadow-lg">
      <pre class="whitespace-pre-line text-green-600 mx-auto">
        <span class="text-blue-400">Type</span><span class="text-white">:</span> ${blockEvent.type}
        <span class="text-blue-400">Timestamp</span><span class="text-white">:</span> ${blockEvent.timestamp}
        <span class="text-blue-400">Block Number</span><span class="text-white">:</span> ${blockEvent.context.blockNumber}
        <span class="text-blue-400">Slot Number</span><span class="text-white">:</span> ${blockEvent.context.slotNumber}
        <span class="text-blue-400">Block Size</span><span class="text-white">:</span> ${blockBodySizeKB.toFixed(2)} KB, ${percentageFull.toFixed(2)}% full
        <span class="text-blue-400">Pool</span><span class="text-white">:</span> <span class="whitespace-pre-line text-wrap" id="issuerVkey" onclick="copyToClipboard(this, 'issuerVkey')">${blockEvent.payload.issuerVkey}</span>
        <span class="text-blue-400">Block Hash</span><span class="text-white">:</span> <span class="whitespace-pre-line text-wrap" id="blockHash" onclick="copyToClipboard(this, 'blockHash')">${blockEvent.payload.blockHash}</span>
        <span class="text-blue-400">Transaction Count</span><span class="text-white">:</span> ${blockEvent.payload.transactionCount}
      </pre>
    </div>
    <div class="text-center py-1 text-white">${timeDifference}</div>
  `;
  // Update the previous slot number for the next message
  prevSlotNumber = blockEvent.context.slotNumber;
}

// Function to display RollbackEvent
function displayRollbackEvent(rollbackEvent, newDiv) {
  // Example HTML content specific to RollbackEvent
  // Calculate time difference
  const timeDifference = calculateTimeDifference(rollbackEvent);
  newDiv.innerHTML = `
    <div class="zoom-in mx-auto max-w-5xl px-4 sm:px-6 lg:px-8 bg-black m-4 p-6 rounded-lg shadow-lg">
      <pre class="mx-auto whitespace-pre-line text-red-600">
        <span class="text-blue-400">Type</span><span class="text-white">:</span> ${rollbackEvent.type}
        <span class="text-blue-400">Timestamp</span><span class="text-white">:</span> ${rollbackEvent.timestamp}                  
        <span class="text-blue-400">Block Hash</span><span class="text-white">:</span> <span class="whitespace-pre-line text-wrap" id="blockHash" onclick="copyToClipboard('${newDiv.id}', 'blockHash')">${rollbackEvent.payload.blockHash}</span>
        <span class="text-blue-400">Slot Number</span><span class="text-white">:</span> ${rollbackEvent.payload.slotNumber}
      </pre>
    </div>
  `;
  // Only display the time difference if it is not an empty string
  if (timeDifference !== "") {
    const timeDifferenceDiv = document.createElement("div");
    timeDifferenceDiv.classList.add("text-center", "py-1", "text-white");
    timeDifferenceDiv.innerText = timeDifference;
    newDiv.appendChild(timeDifferenceDiv);
  }
}

// Function to display TransactionEvent
function displayTransactionEvent(transactionEvent, newDiv) {
  const transactionDiv = document.createElement("div");
  transactionDiv.classList.add("my-2");

  // Calculate time difference
  const timeDifference = calculateTimeDifference(transactionEvent);
  displayedTimeDifference = timeDifference; // Store the displayed time difference

  // Dynamically create the link for each transaction
  const txLink = document.createElement("a");
  const transactionHash = transactionEvent.context.transactionHash;
  const linkId = `tx-link-${transactionHash}`;
  // console.log("Generated Link ID:", linkId); // Log the generated ID
  txLink.href = `#${linkId}`;
  txLink.id = linkId;

  transactionDiv.innerHTML = `
<div class="zoom-in mx-auto max-w-5xl px-4 sm:px-6 lg:px-8 bg-black p-6 rounded-lg shadow-lg relative">
  <div class="mt-2">
    ${txLink.outerHTML}
  </div>
  <pre class="mx-auto whitespace-pre-line text-yellow-600">
    <span class="text-blue-400">Type</span><span class="text-white">:</span> ${transactionEvent.type}
    <span class="text-blue-400">Timestamp</span><span class="text-white">:</span> ${transactionEvent.timestamp}
    <span class="text-blue-400">Block</span><span class="text-white">:</span> ${transactionEvent.context.blockNumber}
    <span class="text-blue-400">Slot</span><span class="text-white">:</span> ${transactionEvent.context.slotNumber}
    <span class="text-blue-400">Tx Hash</span><span class="text-white">:</span> <span class="whitespace-pre-line text-wrap" 
      id="txHash" onclick="copyToClipboard(this)">${transactionEvent.context.transactionHash}</span>
      <span class="text-blue-400">Tx Fee</span><span class="text-white">:</span> ${transactionEvent.payload.fee}
  </pre>
  <div class="absolute top-1 right-1 cursor-pointer" id="arrowButton">
    ➡️ <!-- Right arrow symbol -->
  </div>
</div>
`;

  // Check if the block number has increased
  if (timeDifference !== "") {
    const timeDifferenceDiv = document.createElement("div");
    timeDifferenceDiv.classList.add("text-center", "text-white");
    timeDifferenceDiv.textContent = timeDifference;
    transactionDiv.appendChild(timeDifferenceDiv);
  }

  // Append the new div to the provided container
  newDiv.appendChild(transactionDiv);

  // Add event listener for arrow button click
  const arrowButton = newDiv.querySelector("#arrowButton");
  arrowButton.addEventListener("click", () =>
    handleRightArrowClick(transactionEvent)
  );
}