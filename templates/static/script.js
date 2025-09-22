function Fetch() {
    const keyword = document.getElementById('keyword').value;
    const maxPosts = document.getElementById('maxPosts').value;
    const fetchingDiv = document.getElementById('fetching');

    fetchingDiv.style.display = 'flex'; // show loader

    fetch(`/analyse?keyword=${keyword}&maxPosts=${maxPosts}`)
        .then(res => res.json())
        .then(data => {
            console.log(data); // ✅ handle your data here
            fetchingDiv.style.display = 'none'; // hide loader after success
        })
        .catch(err => {
            console.error(err);
            fetchingDiv.style.display = 'none'; // hide loader on error too
        });
}

async function loadData() {
    try {
        const res = await fetch('/gethistory');
        const json = await res.json();

        // ✅ check "success" at the end
        if (!json.success || !json.data || !Array.isArray(json.data)) {
            document.getElementById('posts-container').innerHTML = "<p>❌ Failed to load data</p>";
            return;
        }

        const data = json.data;
        const buttonContainer = document.getElementById('keyword-buttons');

        // Add "Show All" button
        const showAllBtn = document.createElement('button');
        showAllBtn.textContent = "🌍 Show All";
        showAllBtn.className = "keyword-btn";
        showAllBtn.onclick = () => {
            renderPosts(data);
            renderAnalytics(data);
        };

        buttonContainer.appendChild(showAllBtn);

        // Keyword buttons
        data.forEach(item => {
            if (!item.keyword || item.keyword.trim().length === 0) return; // ✅ skip empty keywords

            const postCount = Array.isArray(item.posts) ? item.posts.length : 0;
            const btn = document.createElement('button');
            btn.textContent = `${item.keyword} (${postCount})`;
            btn.className = "keyword-btn";
            btn.onclick = () => renderPosts([item]);
            buttonContainer.appendChild(btn);
        });

        // Initial render
        renderPosts(data);
        renderAnalytics(data);


    } catch (err) {
        console.log(err)
        document.getElementById('posts-container').innerHTML = `<p>⚠️ Error: ${err.message}</p>`;
    }
}

function renderPosts(dataArray) {
    const postsContainer = document.getElementById('posts-container');

    // Keep analytics div
    const analyticsEl = document.getElementById('analytics');
    postsContainer.innerHTML = '';
    postsContainer.appendChild(analyticsEl);

    // Render posts
    dataArray.forEach(group => {
        if (!group.posts || !Array.isArray(group.posts)) return;
        group.posts.forEach(post => {
            const card = document.createElement('div');
            card.className = 'post-card';
            card.innerHTML = `
            <p><a href="${post.link}" target="_blank">${post.link}</a></p>
            <p>${post.pain_point}</p>
            <div class="meta">
              <span class="tag classification">${post.classification}</span>
              <span class="tag problem">${post.problem_type}</span>
              <span class="tag reoccurrence">Reoccurrence: ${post.reoccurrence}</span>
            </div>
          `;
            postsContainer.appendChild(card);
        });
    });
}

function renderAnalytics(dataArray) {
    const analyticsEl = document.getElementById('analytics');

    let classificationCounts = {};
    let problemCounts = {};

    dataArray.forEach(group => {
        if (!Array.isArray(group.posts)) return;
        group.posts.forEach(post => {
            classificationCounts[post.classification] = (classificationCounts[post.classification] || 0) + 1;
            problemCounts[post.problem_type] = (problemCounts[post.problem_type] || 0) + 1;
        });
    });


    analyticsEl.innerHTML = `
    <div class="analytics-card analytics-classification">
        <h3>Classifications</h3>
        <canvas id="classificationChart"></canvas>
    </div>
    <div class="analytics-card analytics-problem">
        <h3>Problem Types</h3>
        <canvas id="problemChart"></canvas>
    </div>
    `;
    // Classifications chart
    const ctx1 = document.getElementById('classificationChart').getContext('2d');
    new Chart(ctx1, {
        type: 'bar',
        data: {
            labels: Object.keys(classificationCounts),
            datasets: [{
                label: 'Number of Posts',
                data: Object.values(classificationCounts),
                backgroundColor: 'rgba(54, 162, 235, 0.6)',
                borderColor: 'rgba(54, 162, 235, 1)',
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            plugins: { legend: { display: false } },
            scales: { y: { beginAtZero: true } }
        }
    });

    // Problem Types chart
    const ctx2 = document.getElementById('problemChart').getContext('2d');
    new Chart(ctx2, {
        type: 'bar',
        data: {
            labels: Object.keys(problemCounts),
            datasets: [{
                label: 'Number of Posts',
                data: Object.values(problemCounts),
                backgroundColor: 'rgba(255, 99, 132, 0.6)',
                borderColor: 'rgba(255, 99, 132, 1)',
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            plugins: { legend: { display: false } },
            scales: { y: { beginAtZero: true } }
        }
    });
}
