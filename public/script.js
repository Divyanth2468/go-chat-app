let socket = null; // WebSocket connection
let usersocket = null; // WebSocket connection for user list update
let currentFriendId = null;
let receiverId;
let data;
let lastDate = null;

// Inital fetch on page load
async function fetchUserList() {
  try {
    const response = await fetch("/api/userlist"); // Adjust URL if needed
    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }
    data = await response.json();
    // console.log(data);

    // Update Current User
    const User = document.getElementById("current-user");
    User.innerHTML = `
        <div class="current-user-div">
        <img
        src="https://robohash.org/${encodeURIComponent(
          data.UserName
        )}?size=50x50"
        alt="${data.UserName}"
        class="current-user-img"
        onmouseover="this.style.transform='scale(1.1)';" 
        onmouseout="this.style.transform='scale(1)';"
        />
        <div>
        <strong class="current-user-name">${data.UserName}</strong>
        </div>
        </div>
`;

    // Update your UI dynamically
    populateFriendList(data.Friends);
    populateUsersList(data.Users);
    populatePendingRequests(data.Requests);
    setupWebSocket();
  } catch (error) {
    console.error("Error fetching user list:", error);
  }
}

// Example function to populate the friend list
function populateFriendList(friends) {
  if (!friends) return;
  const friendList = document.getElementById("friend-list");
  friendList.innerHTML = ""; // Clear current list

  friends.forEach((friend) => {
    const listItem = document.createElement("li");
    listItem.className = "list-group-item friend-item";
    listItem.innerHTML = `
      <img
          src="https://robohash.org/${encodeURIComponent(
            friend.Name
          )}?size=50x50"
          alt="${friend.Name}"
          style="width: 50px; height: 50px; border-radius: 50%; margin-right: 10px;"
      />
      <strong>${friend.Name}</strong>
  `;
    listItem.onclick = () => openChatWindow(friend.Name, friend.Id, listItem);
    friendList.appendChild(listItem);
  });
}

// Populate Pending Requests
function populatePendingRequests(pendingRequests) {
  const pendingList = document.getElementById("pendingRequestsList");
  pendingList.innerHTML = ""; // Clear current list
  // console.log(pendingList, pendingList.innerHTML);
  if (!pendingRequests) {
    alertDiv = document.createElement("div");
    alertDiv.textContent = "No pending requests";
    alertDiv.className = "alert alert-secondary text-center mt-2";
    alertDiv.id = "EmptyPendingList";
    pendingList.appendChild(alertDiv);
    return;
  }
  // console.log(pendingList, pendingList.innerHTML);

  pendingRequests.forEach((request) => {
    const trimmedName =
      request.Name.length > 15
        ? `${request.Name.substring(0, 12)}...`
        : request.Name;

    const [_, statusPart] = request.Status.split("-"); // Split the status by '-'

    const listItem = document.createElement("li");
    listItem.className =
      "list-group-item pending-item d-flex flex-column align-items-center";

    if (statusPart.trim().toLowerCase() === "requested") {
      // Display grayed-out button for "Requested" status
      listItem.innerHTML = `
          <div class="text-center">
            <strong class="user-name">${trimmedName}</strong>
          </div>
          <button class="btn btn-sm btn-secondary mt-2" disabled>
            Requested
          </button>
        `;
    } else {
      // Display Accept and Reject buttons
      listItem.innerHTML = `
          <div class="text-center">
            <strong class="user-name">${trimmedName}</strong>
          </div>
          <div class="btn-group mt-2">
            <button class="btn btn-sm btn-success accept-request-btn mr-2">
              Accept
            </button>
            <button class="btn btn-sm btn-danger reject-request-btn ml-2">
              Reject
            </button>
          </div>
        `;

      const acceptButton = listItem.querySelector(".accept-request-btn");
      const rejectButton = listItem.querySelector(".reject-request-btn");

      acceptButton.onclick = async () => {
        const res = await handleFriendRequest(request.Id, "accept");
        if (res) {
          listItem.remove(); // Remove the request from the list
        }
      };

      rejectButton.onclick = async () => {
        const res = await handleFriendRequest(request.Id, "reject");
        if (res) {
          listItem.remove(); // Remove the request from the list
        }
      };
    }

    pendingList.appendChild(listItem);
  });
}

// Populate users
function populateUsersList(users) {
  const usersList = document.getElementById("allUsers");
  usersList.innerHTML = ""; // Clear current list

  if (!users) {
    alertDiv = document.createElement("div");
    alertDiv.textContent = "No Users to display";
    alertDiv.className = "alert alert-secondary text-center mt-2";
    alertDiv.id = "EmptyUserList";
    usersList.appendChild(alertDiv);
    return;
  }
  // console.log(users);

  users.forEach((user) => {
    const trimmedName =
      user.Name.length > 15 ? `${user.Name.substring(0, 12)}...` : user.Name;
    const listItem = document.createElement("li");
    listItem.className =
      "list-group-item user-item d-flex justify-content-between align-items-center";

    listItem.innerHTML = `
      <div class="d-flex align-items-center">
        <strong>${trimmedName}</strong>
      </div>
      <button class="btn btn-sm btn-primary add-friend-btn">
        Add Friend
      </button>
    `;

    const addFriendButton = listItem.querySelector(".add-friend-btn");
    addFriendButton.onclick = () => {
      const res = handleFriendRequest(user.Id, "");
      if (res) {
        // Update button text and style
        addFriendButton.textContent = "Request Sent";
        addFriendButton.classList.remove("btn-primary");
        addFriendButton.classList.add("btn-success");

        // Disable the button
        addFriendButton.disabled = true;

        // Remove the list item after 2 seconds
        setTimeout(() => {
          listItem.remove();
        }, 2000);
      }
    };
    usersList.appendChild(listItem);
  });
}

// Socket for real time list updates
function setupWebSocket() {
  // Close any existing WebSocket connection
  if (usersocket) {
    usersocket.close();
  }

  // Establish a new WebSocket connection
  usersocket = new WebSocket(
    `ws://localhost:8080/ws-friend-request?user_id=${data.UserId}`
  );

  usersocket.onopen = () => {
    console.log("WebSocket connected for Friend List Updates");
  };

  usersocket.onmessage = (event) => {
    const data = JSON.parse(event.data);
    // console.log(data);
    // Handle updates for the pending friend list
    const updatedPendingList = data.RequestList;
    populatePendingRequests(updatedPendingList);

    // Handle updates for the accepted/rejected friend list
    const updatedFriendList = data.FriendList;
    populateFriendList(updatedFriendList);

    // Handle Updates for user list
    const updatedUserList = data.UserList;
    populateUsersList(updatedUserList);
  };

  usersocket.onclose = () => {
    console.log("WebSocket connection closed");
  };
}

// Function to send requests
async function handleFriendRequest(reqUserId, action) {
  if (!usersocket) {
    console.error("WebSocket is not initialized");
    return;
  }

  // Send friend request action via WebSocket
  usersocket.send(
    JSON.stringify({
      user_id: data.UserId,
      req_user_Id: reqUserId,
      action: action,
    })
  );
}

// Open chat window for chat
async function openChatWindow(friendName, friendId, friendElement) {
  lastDate = null;
  if (currentFriendId !== friendId) {
    try {
      const response = await fetch(
        `/api/messages?senderId=${data.UserId}&friendId=${friendId}`
      );
      if (!response.ok) throw new Error("Error fetching messages");
      const messages = await response.json();
      if (messages != null) {
        displayMessages(messages);
      } else {
        document.getElementById("chat-messages").innerHTML = "";
        displayMessage("");
      }
    } catch (err) {
      console.log("Error fetching messages: ", err);
    }

    if (socket) {
      socket.close();
    }

    socket = new WebSocket(
      `ws://localhost:8080/ws?senderId=${data.UserId}&friendId=${friendId}`
    );

    socket.onopen = () =>
      console.log("WebSocket connected for friend:", friendName);

    socket.onmessage = (event) => {
      const message = JSON.parse(event.data).Mes;
      displayMessage(message);
    };

    socket.onclose = () => console.log("WebSocket connection closed");

    currentFriendId = friendId;
  }

  const activeItem = document.querySelector(".friend-item.active");
  if (activeItem) activeItem.classList.remove("active");
  friendElement.classList.add("active");

  document.getElementById("friend-name-text").textContent = friendName;
  document.getElementById("chat-interface").style.display = "block";
}

// Function to display all messages in the chat history
function displayMessages(messages) {
  // Clear the existing messages in the chat container (optional)
  const chatContainer = document.getElementById("chat-messages");
  chatContainer.innerHTML = "";

  // Loop through each message and display it
  // console.log(messages);
  messages.forEach((message) => {
    displayMessage(message); // Call displayMessage for each individual message
  });
}

// Send messages
function sendMessage() {
  const messageInput = document.getElementById("chat-message");
  const message = messageInput.value.trim();
  // console.log(message);
  if (message && socket && socket.readyState === WebSocket.OPEN) {
    socket.send(message);
  }
  messageInput.value = "";
}

// Display messages
function displayMessage(message) {
  // Hide the placeholder
  const placeholder = document.getElementById("placeholder-message");
  placeholder.style.display = "none";

  // Set chat interface to be visible
  const chatInterface = document.getElementById("chat-interface");
  chatInterface.style.display = "flex !important";
  // console.log("Displaying", message.SenderId, message.ReceiverId);

  if (!message) return;

  const chatWindow = document.getElementById("chat-messages");
  const currentDate = formatDate(message.TimeStamp);

  // Check if new date header needs to be added
  if (lastDate != currentDate) {
    // Create New Date Header
    const dateDiv = document.createElement("div");
    dateDiv.classList.add("date-header");
    dateDiv.textContent = currentDate;

    // Append the date header to the chat window
    chatWindow.appendChild(dateDiv);

    // Update the lastDate
    lastDate = currentDate;
  }

  const messageDiv = document.createElement("div");
  messageDiv.classList.add("message");

  // Assign class based on sender
  if (message.SenderId === data.UserId) {
    messageDiv.classList.add("receiver");
  } else {
    messageDiv.classList.add("sender");
  }

  // Create message container
  const messageContainer = document.createElement("p");
  messageContainer.classList.add("message-container");

  // Create message content (text of the message)
  const messageContent = document.createElement("span");
  messageContent.classList.add("message-content");
  messageContent.textContent = message.Message;
  messageContainer.appendChild(messageContent);

  // Create the timestamp span element
  const timestampSpan = document.createElement("span");
  timestampSpan.classList.add("message-time");
  timestampSpan.textContent = formatTimestamp(message.TimeStamp); // Format your timestamp as needed
  messageContainer.appendChild(timestampSpan);

  // Append message content and timestamp to the message div
  messageDiv.appendChild(messageContainer);

  // Append the message div to the chat window
  chatWindow.appendChild(messageDiv);
  scrollToBottom();
}

// Time conversions
function formatTimestamp(utcTimestamp) {
  const date = new Date(utcTimestamp); // Create a Date object from UTC timestamp
  const hours = date.getHours().toString().padStart(2, "0"); // Get hours in UTC and pad with leading zeros if needed
  const minutes = date.getMinutes().toString().padStart(2, "0"); // Get minutes in UTC and pad with leading zeros if needed
  return `${hours}:${minutes}`; // Return in "HH:MM" format
}

// Date conversions
function formatDate(utcTimestamp) {
  const date = new Date(utcTimestamp); // Create a Date object from UTC timestamp
  const now = new Date(); // Current date and time

  // Calculate the difference in days
  const diffTime = now - date; // Difference in milliseconds
  const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24)); // Convert to days

  if (diffDays === 0) {
    return "Today";
  } else if (diffDays === 1) {
    return "Yesterday";
  } else {
    // For older dates, format as YYYY-MM-DD
    const displayDate =
      date.getFullYear() +
      "-" +
      String(date.getMonth() + 1).padStart(2, "0") + // Ensure two digits
      "-" +
      String(date.getDate()).padStart(2, "0"); // Ensure two digits
    return displayDate;
  }
}

document
  .getElementById("send-message-btn")
  .addEventListener("click", sendMessage);

// Function to scroll to the bottom of the chat container
function scrollToBottom() {
  const chatMessages = document.getElementById("chat-messages");
  chatMessages.scrollTop = chatMessages.scrollHeight;
}

window.onload = () => {
  // Call the function on page load or when needed
  fetchUserList();
};

// `wss://go-chat-app-production-60eb.up.railway.app/ws?senderId=${data.UserId}&friendId=${friendId}`
