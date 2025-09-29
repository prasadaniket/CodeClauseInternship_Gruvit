import TrackPlayer, { Capability, RepeatMode } from 'react-native-track-player';

export async function initializePlayer() {
  try {
    await TrackPlayer.setupPlayer({
      waitForBuffer: true,
    });

    await TrackPlayer.updateOptions({
      capabilities: [
        Capability.Play,
        Capability.Pause,
        Capability.SkipToNext,
        Capability.SkipToPrevious,
        Capability.Stop,
        Capability.SeekTo,
      ],
      compactCapabilities: [
        Capability.Play,
        Capability.Pause,
        Capability.SkipToNext,
        Capability.SkipToPrevious,
      ],
      notificationCapabilities: [
        Capability.Play,
        Capability.Pause,
        Capability.SkipToNext,
        Capability.SkipToPrevious,
      ],
    });

    console.log('TrackPlayer initialized successfully');
  } catch (error) {
    console.error('Failed to initialize TrackPlayer:', error);
  }
}

export async function playTrack(track: any) {
  try {
    await TrackPlayer.add(track);
    await TrackPlayer.play();
  } catch (error) {
    console.error('Failed to play track:', error);
    throw error;
  }
}

export async function pauseTrack() {
  try {
    await TrackPlayer.pause();
  } catch (error) {
    console.error('Failed to pause track:', error);
    throw error;
  }
}

export async function resumeTrack() {
  try {
    await TrackPlayer.play();
  } catch (error) {
    console.error('Failed to resume track:', error);
    throw error;
  }
}

export async function skipToNext() {
  try {
    await TrackPlayer.skipToNext();
  } catch (error) {
    console.error('Failed to skip to next:', error);
    throw error;
  }
}

export async function skipToPrevious() {
  try {
    await TrackPlayer.skipToPrevious();
  } catch (error) {
    console.error('Failed to skip to previous:', error);
    throw error;
  }
}

export async function seekTo(position: number) {
  try {
    await TrackPlayer.seekTo(position);
  } catch (error) {
    console.error('Failed to seek:', error);
    throw error;
  }
}

export async function setRepeatMode(mode: RepeatMode) {
  try {
    await TrackPlayer.setRepeatMode(mode);
  } catch (error) {
    console.error('Failed to set repeat mode:', error);
    throw error;
  }
}

export async function setQueue(tracks: any[]) {
  try {
    await TrackPlayer.reset();
    await TrackPlayer.add(tracks);
  } catch (error) {
    console.error('Failed to set queue:', error);
    throw error;
  }
}

export async function addToQueue(track: any) {
  try {
    await TrackPlayer.add(track);
  } catch (error) {
    console.error('Failed to add to queue:', error);
    throw error;
  }
}

export async function removeFromQueue(index: number) {
  try {
    await TrackPlayer.remove(index);
  } catch (error) {
    console.error('Failed to remove from queue:', error);
    throw error;
  }
}

export async function getCurrentTrack() {
  try {
    return await TrackPlayer.getCurrentTrack();
  } catch (error) {
    console.error('Failed to get current track:', error);
    return null;
  }
}

export async function getPosition() {
  try {
    return await TrackPlayer.getPosition();
  } catch (error) {
    console.error('Failed to get position:', error);
    return 0;
  }
}

export async function getDuration() {
  try {
    return await TrackPlayer.getDuration();
  } catch (error) {
    console.error('Failed to get duration:', error);
    return 0;
  }
}

export async function getState() {
  try {
    return await TrackPlayer.getState();
  } catch (error) {
    console.error('Failed to get state:', error);
    return null;
  }
}
