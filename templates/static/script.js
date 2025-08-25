function Fetch() {
    const keyword = document.getElementById('keyword').value;
    const maxPosts = document.getElementById('maxPosts').value;
    const resultsDiv = document.getElementById('results');
    const fetchingDiv = document.getElementById('fetching');

    resultsDiv.innerHTML = '';
    fetchingDiv.style.display = 'block'; // show animation

    if (window.evtSource) window.evtSource.close();

    const params = new URLSearchParams({ keyword, max: maxPosts });
    window.evtSource = new EventSource(`/analyse?${params.toString()}`);

    window.evtSource.onmessage = function (event) {
        try {
            const data = JSON.parse(event.data);
            const card = document.createElement('div');
            card.className = 'card';
            card.innerHTML = `
        <div class="community">Community: ${data.Community}</div>
        <div class="summary">${data.Summary}</div>
        <div class="solution"><span>Proposed Solution:</span> ${data.MicrosaasSolution}</div>
      `;
            resultsDiv.appendChild(card);
        } catch (err) {
            console.error("Failed to parse SSE data:", err, event.data);
        }
    };

    // Listen for "done" event sent from backend
    window.evtSource.addEventListener('done', () => {
        fetchingDiv.style.display = 'none';
        window.evtSource.close();
    });

    window.evtSource.onerror = function (err) {
        console.error("SSE error", err);
        fetchingDiv.style.display = 'none';
        window.evtSource.close();
    };
}

// Stop manually
function StopFetching() {
    if (window.evtSource) {
        window.evtSource.close();
        document.getElementById('fetching').style.display = 'none';
    }
}

function getRatingColor(rating) {
    if (rating <= 0.9) return '#FFFF00';        // yellow
    if (rating <= 0.95) return '#FFA500';       // orange
    return '#FF4500';                            // red for highest severity
}

function getHistory() {
    fetch('/gethistory')
        .then(res => res.json())
        .then(data => {
            const postsTable = document.getElementById('postsTable');
            const postsBody = document.getElementById('postsBody');
            const loader = document.getElementById('loader');

            if (data.success && data.data.length) {
                data.data.forEach(keywordGroup => {
                    keywordGroup.posts.forEach(post => {
                        if (post.Rating < 0.9) return;

                        const tr = document.createElement('tr');
                        const ratingColor = getRatingColor(post.Rating);

                        tr.innerHTML = `
                            <td>${post.Community}</td>
                            <td>${post.Summary}</td>
                            <td>${post.MicrosaasSolution}</td>
                            <td style="background-color:${ratingColor}; font-weight:bold; text-align:center;">${post.Rating}</td>
                            <td>
                                <button onclick="window.open('${post.Link}', '_blank')">🏃‍➡️</button>
                            </td>`;
                        postsBody.appendChild(tr);
                    });
                });

                loader.style.display = 'none';
                postsTable.style.display = '';
                document.getElementById('filter').style.display = 'none';
            }
            else {
                loader.innerText = 'No posts found.';
            }
        })
        .catch(err => {
            loader.innerText = 'Error loading posts.';
            console.error(err);
        });
}
