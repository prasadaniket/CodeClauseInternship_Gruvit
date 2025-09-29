import React, { createContext, useContext, useReducer, useEffect } from 'react';
import TrackPlayer, { 
  Capability, 
  State, 
  Event, 
  RepeatMode,
  Track,
  useTrackPlayerEvents,
} from 'react-native-track-player';

// Types
interface MusicPlayerState {
  currentTrack: Track | null;
  isPlaying: boolean;
  position: number;
  duration: number;
  repeatMode: RepeatMode;
  shuffleMode: boolean;
  queue: Track[];
  currentIndex: number;
  isLoading: boolean;
  error: string | null;
}

type MusicPlayerAction =
  | { type: 'SET_TRACK'; payload: Track }
  | { type: 'PLAY' }
  | { type: 'PAUSE' }
  | { type: 'TOGGLE_PLAY_PAUSE' }
  | { type: 'SET_POSITION'; payload: number }
  | { type: 'SET_DURATION'; payload: number }
  | { type: 'SET_REPEAT_MODE'; payload: RepeatMode }
  | { type: 'TOGGLE_SHUFFLE' }
  | { type: 'NEXT_TRACK' }
  | { type: 'PREVIOUS_TRACK' }
  | { type: 'SET_QUEUE'; payload: Track[] }
  | { type: 'ADD_TO_QUEUE'; payload: Track }
  | { type: 'REMOVE_FROM_QUEUE'; payload: number }
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_ERROR'; payload: string | null }
  | { type: 'CLEAR_ERROR' };

// Initial state
const initialState: MusicPlayerState = {
  currentTrack: null,
  isPlaying: false,
  position: 0,
  duration: 0,
  repeatMode: RepeatMode.Off,
  shuffleMode: false,
  queue: [],
  currentIndex: -1,
  isLoading: false,
  error: null,
};

// Reducer
function musicPlayerReducer(state: MusicPlayerState, action: MusicPlayerAction): MusicPlayerState {
  switch (action.type) {
    case 'SET_TRACK':
      return {
        ...state,
        currentTrack: action.payload,
        position: 0,
        error: null,
      };

    case 'PLAY':
      return {
        ...state,
        isPlaying: true,
        error: null,
      };

    case 'PAUSE':
      return {
        ...state,
        isPlaying: false,
      };

    case 'TOGGLE_PLAY_PAUSE':
      return {
        ...state,
        isPlaying: !state.isPlaying,
        error: null,
      };

    case 'SET_POSITION':
      return {
        ...state,
        position: action.payload,
      };

    case 'SET_DURATION':
      return {
        ...state,
        duration: action.payload,
      };

    case 'SET_REPEAT_MODE':
      return {
        ...state,
        repeatMode: action.payload,
      };

    case 'TOGGLE_SHUFFLE':
      return {
        ...state,
        shuffleMode: !state.shuffleMode,
      };

    case 'NEXT_TRACK':
      if (state.queue.length === 0) return state;
      
      let nextIndex = state.currentIndex + 1;
      if (nextIndex >= state.queue.length) {
        if (state.repeatMode === RepeatMode.Queue) {
          nextIndex = 0;
        } else {
          return { ...state, isPlaying: false };
        }
      }
      
      return {
        ...state,
        currentIndex: nextIndex,
        currentTrack: state.queue[nextIndex],
        position: 0,
      };

    case 'PREVIOUS_TRACK':
      if (state.queue.length === 0) return state;
      
      let prevIndex = state.currentIndex - 1;
      if (prevIndex < 0) {
        if (state.repeatMode === RepeatMode.Queue) {
          prevIndex = state.queue.length - 1;
        } else {
          return state;
        }
      }
      
      return {
        ...state,
        currentIndex: prevIndex,
        currentTrack: state.queue[prevIndex],
        position: 0,
      };

    case 'SET_QUEUE':
      return {
        ...state,
        queue: action.payload,
        currentIndex: action.payload.length > 0 ? 0 : -1,
        currentTrack: action.payload.length > 0 ? action.payload[0] : null,
        position: 0,
      };

    case 'ADD_TO_QUEUE':
      const newQueue = [...state.queue, action.payload];
      return {
        ...state,
        queue: newQueue,
        currentIndex: state.currentIndex === -1 ? 0 : state.currentIndex,
        currentTrack: state.currentTrack || action.payload,
      };

    case 'REMOVE_FROM_QUEUE':
      const filteredQueue = state.queue.filter((_, index) => index !== action.payload);
      let newCurrentIndex = state.currentIndex;
      
      if (action.payload < state.currentIndex) {
        newCurrentIndex = state.currentIndex - 1;
      } else if (action.payload === state.currentIndex) {
        if (filteredQueue.length === 0) {
          newCurrentIndex = -1;
        } else if (newCurrentIndex >= filteredQueue.length) {
          newCurrentIndex = filteredQueue.length - 1;
        }
      }
      
      return {
        ...state,
        queue: filteredQueue,
        currentIndex: newCurrentIndex,
        currentTrack: filteredQueue.length > 0 && newCurrentIndex >= 0 ? filteredQueue[newCurrentIndex] : null,
      };

    case 'SET_LOADING':
      return {
        ...state,
        isLoading: action.payload,
      };

    case 'SET_ERROR':
      return {
        ...state,
        error: action.payload,
        isLoading: false,
      };

    case 'CLEAR_ERROR':
      return {
        ...state,
        error: null,
      };

    default:
      return state;
  }
}

// Context
const MusicPlayerContext = createContext<{
  state: MusicPlayerState;
  dispatch: React.Dispatch<MusicPlayerAction>;
} | null>(null);

// Provider
export function MusicPlayerProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(musicPlayerReducer, initialState);

  // Track player events
  useTrackPlayerEvents([Event.PlaybackState, Event.PlaybackTrackChanged], (event) => {
    switch (event.type) {
      case Event.PlaybackState:
        if (event.state === State.Playing) {
          dispatch({ type: 'PLAY' });
        } else if (event.state === State.Paused) {
          dispatch({ type: 'PAUSE' });
        }
        break;
      case Event.PlaybackTrackChanged:
        if (event.track) {
          dispatch({ type: 'SET_TRACK', payload: event.track });
        }
        break;
    }
  });

  // Update position periodically
  useEffect(() => {
    const interval = setInterval(async () => {
      try {
        const position = await TrackPlayer.getPosition();
        dispatch({ type: 'SET_POSITION', payload: position });
      } catch (error) {
        console.error('Failed to get position:', error);
      }
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  return (
    <MusicPlayerContext.Provider value={{ state, dispatch }}>
      {children}
    </MusicPlayerContext.Provider>
  );
}

// Hook
export function useMusicPlayer() {
  const context = useContext(MusicPlayerContext);
  if (!context) {
    throw new Error('useMusicPlayer must be used within a MusicPlayerProvider');
  }
  return context;
}

// Helper functions
export const musicPlayerActions = {
  setTrack: (track: Track) => ({ type: 'SET_TRACK' as const, payload: track }),
  play: () => ({ type: 'PLAY' as const }),
  pause: () => ({ type: 'PAUSE' as const }),
  togglePlayPause: () => ({ type: 'TOGGLE_PLAY_PAUSE' as const }),
  setPosition: (position: number) => ({ type: 'SET_POSITION' as const, payload: position }),
  setDuration: (duration: number) => ({ type: 'SET_DURATION' as const, payload: duration }),
  setRepeatMode: (mode: RepeatMode) => ({ type: 'SET_REPEAT_MODE' as const, payload: mode }),
  toggleShuffle: () => ({ type: 'TOGGLE_SHUFFLE' as const }),
  nextTrack: () => ({ type: 'NEXT_TRACK' as const }),
  previousTrack: () => ({ type: 'PREVIOUS_TRACK' as const }),
  setQueue: (tracks: Track[]) => ({ type: 'SET_QUEUE' as const, payload: tracks }),
  addToQueue: (track: Track) => ({ type: 'ADD_TO_QUEUE' as const, payload: track }),
  removeFromQueue: (index: number) => ({ type: 'REMOVE_FROM_QUEUE' as const, payload: index }),
  setLoading: (loading: boolean) => ({ type: 'SET_LOADING' as const, payload: loading }),
  setError: (error: string | null) => ({ type: 'SET_ERROR' as const, payload: error }),
  clearError: () => ({ type: 'CLEAR_ERROR' as const }),
};
