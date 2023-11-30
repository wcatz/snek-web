let prevSlotNumber

function copyToClipboard(element, property) {
  const propertySpan = element;
  const tempInput = document.createElement("textarea");
  tempInput.value = propertySpan.innerText;
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

  // Close menu when clicking outside the menu or pressing the Escape key
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
// Determine the slot number based on the event type
const currentSlotNumber =
  currentEvent.type === "chainsync.block"
    ? currentEvent.context.slotNumber
    : currentEvent.type === "chainsync.rollback"
    ? currentEvent.payload.slotNumber
    : currentEvent.context.slotNumber;

// Only calculate time difference if the slot number has increased
if (currentSlotNumber > prevSlotNumber) {
  // Calculate time difference based on the slot number of the current event
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
  : currentEvent.context.slotNumber;

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
      <span class="text-blue-400">Pool</span><span class="text-white">:</span> <span class="whitespace-pre-line text-wrap" 
      id="issuerVkey" onclick="copyToClipboard(this, 'issuerVkey')">${
        blockEvent.payload.issuerVkey
      }</span>
      <span class="text-blue-400">Block Hash</span><span class="text-white">:</span> <span class="whitespace-pre-line text-wrap" 
      id="blockHash" onclick="copyToClipboard(this, 'blockHash')">${
        blockEvent.payload.blockHash
      }</span>
      <span class="text-blue-400">Transaction Count</span><span class="text-white">:</span> ${
        blockEvent.payload.transactionCount
      }
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
<span class="text-blue-400">Block Hash</span><span class="text-white">:</span> <span class="whitespace-pre-line text-wrap" 
id="blockHash" onclick="copyToClipboard('${newDiv.id}', 'blockHash')">${rollbackEvent.payload.blockHash}</span>
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
