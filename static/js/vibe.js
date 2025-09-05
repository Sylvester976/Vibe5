let currentStory = 0;
let isPlaying = true;
let storyTimer;
let progressTimer;
let timeRemaining = 30;

function populateData() {
    // Populate tracks
    const tracksContainer = document.getElementById('tracks-container');
    vibeData.TopTracks.forEach((track, index) => {
        const trackDiv = document.createElement('div');
        trackDiv.className = 'track-item d-flex align-items-center';
        trackDiv.innerHTML = `
            <span class="track-number">${index + 1}</span>
            <div class="track-info">
                <h5>${track.Name}</h5>
                <p>${track.Artist} • ${track.Album}</p>
            </div>
        `;
        tracksContainer.appendChild(trackDiv);
    });

    // Populate artists
    const artistsContainer = document.getElementById('artists-container');
    vibeData.TopArtists.forEach((artist, index) => {
        const artistDiv = document.createElement('div');
        artistDiv.className = 'artist-item d-flex align-items-center';
        artistDiv.innerHTML = `
            <span class="track-number">${index + 1}</span>
            <div class="track-info">
                <h5>${artist.Name}</h5>
                <p>${artist.Genres.join(' • ')}</p>
            </div>
        `;
        artistsContainer.appendChild(artistDiv);
    });

    // Populate genres
    const genresContainer = document.getElementById('genres-container');
    vibeData.TopGenres.forEach((genre) => {
        const genreDiv = document.createElement('div');
        genreDiv.className = 'genre-tag';
        genreDiv.textContent = genre.charAt(0).toUpperCase() + genre.slice(1);
        genresContainer.appendChild(genreDiv);
    });
}

function startStoryTimer() {
    if (!isPlaying) return;

    timeRemaining = 30;
    updateTimer();

    storyTimer = setTimeout(() => {
        nextStory();
    }, 30000);

    const progressFill = document.getElementById(`progress-${currentStory}`);
    if (progressFill) {
        progressFill.style.width = '0%';
        setTimeout(() => {
            progressFill.style.transition = 'width 60s linear';
            progressFill.style.width = '100%';
        }, 100);
    }

    progressTimer = setInterval(() => {
        timeRemaining--;
        updateTimer();
        if (timeRemaining <= 0) clearInterval(progressTimer);
    }, 1000);
}

function updateTimer() {
    document.getElementById('timer').textContent = timeRemaining;
}

function stopTimers() {
    clearTimeout(storyTimer);
    clearInterval(progressTimer);

    const progressFill = document.getElementById(`progress-${currentStory}`);
    if (progressFill) {
        progressFill.style.transition = 'width 0.1s ease';
    }
}

function showStory(index) {
    const stories = document.querySelectorAll('.story');

    stories.forEach((story, i) => {
        story.classList.remove('active', 'prev');
        if (i < index) story.classList.add('prev');
        else if (i === index) story.classList.add('active');
    });

    const progressBars = document.querySelectorAll('.progress-fill');
    progressBars.forEach((bar, i) => {
        if (i < index) {
            bar.style.width = '100%';
            bar.style.transition = 'width 0.3s ease';
        } else {
            bar.style.width = '0%';
            bar.style.transition = 'width 0.1s ease';
        }
    });
}

function nextStory() {
    stopTimers();
    if (currentStory < 4) {
        currentStory++;
        showStory(currentStory);
        if (isPlaying) startStoryTimer();
    }
}

function previousStory() {
    stopTimers();
    if (currentStory > 0) {
        currentStory--;
        showStory(currentStory);
        if (isPlaying) startStoryTimer();
    }
}

function togglePlayPause() {
    isPlaying = !isPlaying;
    const playPauseBtn = document.getElementById('playPauseBtn');

    if (isPlaying) {
        playPauseBtn.innerHTML = '<i class="fas fa-pause"></i>';
        startStoryTimer();
    } else {
        playPauseBtn.innerHTML = '<i class="fas fa-play"></i>';
        stopTimers();
    }
}

// Keyboard controls
document.addEventListener('keydown', (e) => {
    switch (e.key) {
        case 'ArrowLeft':
            previousStory();
            break;
        case 'ArrowRight':
        case ' ':
            e.preventDefault();
            nextStory();
            break;
        case 'Enter':
            togglePlayPause();
            break;
    }
});

// Touch controls
let touchStartX = 0, touchEndX = 0;
document.addEventListener('touchstart', (e) => {
    touchStartX = e.changedTouches[0].screenX;
});
document.addEventListener('touchend', (e) => {
    touchEndX = e.changedTouches[0].screenX;
    handleSwipe();
});
function handleSwipe() {
    const diff = touchStartX - touchEndX;
    if (Math.abs(diff) > 50) {
        if (diff > 0) nextStory(); else previousStory();
    }
}

// Click-to-advance
document.addEventListener('click', (e) => {
    if (!e.target.closest('.controls') && !e.target.closest('button')) {
        const clickX = e.clientX;
        const windowWidth = window.innerWidth;
        if (clickX > windowWidth / 2) nextStory();
        else previousStory();
    }
});

// Initialize
populateData();
showStory(0);
startStoryTimer();
