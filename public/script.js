let socket = null; // WebSocket connection
let currentFriendId = null;
let receiverId;
let data;

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
  } catch (error) {
    console.error("Error fetching user list:", error);
  }
}

// Example function to populate the friend list
function populateFriendList(friends) {
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

// Call the function on page load or when needed
fetchUserList();

async function openChatWindow(friendName, friendId, friendElement) {
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
      `wss://go-chat-app-production-60eb.up.railway.app/ws?senderId=${data.UserId}&friendId=${friendId}`
    );

    socket.onopen = () =>
      console.log("WebSocket connected for friend:", friendName);

    socket.onmessage = (event) => {
      const message = JSON.parse(event.data).Mes;
      // message.Timestamp = new Date().toISOString();
      // console.log(message.TimeStamp);
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

function sendMessage() {
  const messageInput = document.getElementById("chat-message");
  const message = messageInput.value.trim();

  if (message && socket && socket.readyState === WebSocket.OPEN) {
    socket.send(message);
  }
  messageInput.value = "";
}

function displayMessage(message) {
  // Hide the placeholder
  const placeholder = document.getElementById("placeholder-message");
  placeholder.style.display = "none";

  // Set chat interface to be visible
  const chatInterface = document.getElementById("chat-interface");
  chatInterface.style.display = "flex !important";
  // console.log("Displaying", message.SenderId, message.ReceiverId);

  if (!message) return;
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
  // messageContent.appendChild(timestampSpan);

  // Append the message div to the chat window
  document.getElementById("chat-messages").appendChild(messageDiv);
  scrollToBottom();
}

// Time conversions
function formatTimestamp(utcTimestamp) {
  const date = new Date(utcTimestamp); // Create a Date object from UTC timestamp
  const hours = date.getHours().toString().padStart(2, "0"); // Get hours in UTC and pad with leading zeros if needed
  const minutes = date.getMinutes().toString().padStart(2, "0"); // Get minutes in UTC and pad with leading zeros if needed
  return `${hours}:${minutes}`; // Return in "HH:MM" format
}

document
  .getElementById("send-message-btn")
  .addEventListener("click", sendMessage);

// Function to scroll to the bottom of the chat container
function scrollToBottom() {
  const chatMessages = document.getElementById("chat-messages");
  chatMessages.scrollTop = chatMessages.scrollHeight;
}
