import React, { useEffect, useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  StyleSheet,
  TouchableOpacity,
  Image,
  RefreshControl,
  Dimensions,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import Icon from 'react-native-vector-icons/MaterialIcons';
import LinearGradient from 'react-native-linear-gradient';

const { width } = Dimensions.get('window');

interface Track {
  id: string;
  title: string;
  artist: string;
  image_url: string;
  duration: string;
}

interface Playlist {
  id: string;
  name: string;
  image_url: string;
  track_count: number;
}

export default function HomeScreen({ navigation }: any) {
  const [tracks, setTracks] = useState<Track[]>([]);
  const [playlists, setPlaylists] = useState<Playlist[]>([]);
  const [refreshing, setRefreshing] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      // Simulate API calls
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // Mock data
      setTracks([
        {
          id: '1',
          title: 'Sample Track 1',
          artist: 'Sample Artist',
          image_url: 'https://via.placeholder.com/150',
          duration: '3:45',
        },
        {
          id: '2',
          title: 'Sample Track 2',
          artist: 'Sample Artist 2',
          image_url: 'https://via.placeholder.com/150',
          duration: '4:12',
        },
      ]);

      setPlaylists([
        {
          id: '1',
          name: 'My Favorites',
          image_url: 'https://via.placeholder.com/150',
          track_count: 25,
        },
        {
          id: '2',
          name: 'Workout Mix',
          image_url: 'https://via.placeholder.com/150',
          track_count: 15,
        },
      ]);
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setLoading(false);
    }
  };

  const onRefresh = async () => {
    setRefreshing(true);
    await loadData();
    setRefreshing(false);
  };

  const renderTrack = (track: Track) => (
    <TouchableOpacity
      key={track.id}
      style={styles.trackItem}
      onPress={() => {
        // Navigate to player or play track
        navigation.navigate('Player', { track });
      }}
    >
      <Image source={{ uri: track.image_url }} style={styles.trackImage} />
      <View style={styles.trackInfo}>
        <Text style={styles.trackTitle} numberOfLines={1}>
          {track.title}
        </Text>
        <Text style={styles.trackArtist} numberOfLines={1}>
          {track.artist}
        </Text>
      </View>
      <View style={styles.trackActions}>
        <Text style={styles.trackDuration}>{track.duration}</Text>
        <TouchableOpacity style={styles.playButton}>
          <Icon name="play-arrow" size={24} color="#7c3aed" />
        </TouchableOpacity>
      </View>
    </TouchableOpacity>
  );

  const renderPlaylist = (playlist: Playlist) => (
    <TouchableOpacity
      key={playlist.id}
      style={styles.playlistItem}
      onPress={() => {
        // Navigate to playlist
        navigation.navigate('Playlist', { playlist });
      }}
    >
      <Image source={{ uri: playlist.image_url }} style={styles.playlistImage} />
      <View style={styles.playlistInfo}>
        <Text style={styles.playlistName} numberOfLines={1}>
          {playlist.name}
        </Text>
        <Text style={styles.playlistCount}>
          {playlist.track_count} tracks
        </Text>
      </View>
    </TouchableOpacity>
  );

  if (loading) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.loadingContainer}>
          <Text style={styles.loadingText}>Loading...</Text>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView
        style={styles.scrollView}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
      >
        {/* Header */}
        <LinearGradient
          colors={['#7c3aed', '#a855f7']}
          style={styles.header}
        >
          <View style={styles.headerContent}>
            <Text style={styles.headerTitle}>Welcome to Gruvit</Text>
            <Text style={styles.headerSubtitle}>Discover your next favorite song</Text>
          </View>
        </LinearGradient>

        {/* Quick Actions */}
        <View style={styles.quickActions}>
          <TouchableOpacity style={styles.quickActionButton}>
            <Icon name="search" size={24} color="#7c3aed" />
            <Text style={styles.quickActionText}>Search</Text>
          </TouchableOpacity>
          <TouchableOpacity style={styles.quickActionButton}>
            <Icon name="shuffle" size={24} color="#7c3aed" />
            <Text style={styles.quickActionText}>Shuffle</Text>
          </TouchableOpacity>
          <TouchableOpacity style={styles.quickActionButton}>
            <Icon name="favorite" size={24} color="#7c3aed" />
            <Text style={styles.quickActionText}>Favorites</Text>
          </TouchableOpacity>
        </View>

        {/* Recent Tracks */}
        <View style={styles.section}>
          <View style={styles.sectionHeader}>
            <Text style={styles.sectionTitle}>Recent Tracks</Text>
            <TouchableOpacity>
              <Text style={styles.sectionAction}>See All</Text>
            </TouchableOpacity>
          </View>
          {tracks.map(renderTrack)}
        </View>

        {/* My Playlists */}
        <View style={styles.section}>
          <View style={styles.sectionHeader}>
            <Text style={styles.sectionTitle}>My Playlists</Text>
            <TouchableOpacity>
              <Text style={styles.sectionAction}>See All</Text>
            </TouchableOpacity>
          </View>
          <ScrollView horizontal showsHorizontalScrollIndicator={false}>
            {playlists.map(renderPlaylist)}
          </ScrollView>
        </View>

        {/* Trending */}
        <View style={styles.section}>
          <View style={styles.sectionHeader}>
            <Text style={styles.sectionTitle}>Trending Now</Text>
            <TouchableOpacity>
              <Text style={styles.sectionAction}>See All</Text>
            </TouchableOpacity>
          </View>
          {tracks.map(renderTrack)}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f8fafc',
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingText: {
    fontSize: 16,
    color: '#6b7280',
  },
  scrollView: {
    flex: 1,
  },
  header: {
    padding: 20,
    marginBottom: 20,
  },
  headerContent: {
    alignItems: 'center',
  },
  headerTitle: {
    fontSize: 24,
    fontWeight: 'bold',
    color: 'white',
    marginBottom: 8,
  },
  headerSubtitle: {
    fontSize: 16,
    color: 'rgba(255, 255, 255, 0.8)',
  },
  quickActions: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    paddingHorizontal: 20,
    marginBottom: 20,
  },
  quickActionButton: {
    alignItems: 'center',
    padding: 16,
    backgroundColor: 'white',
    borderRadius: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  quickActionText: {
    marginTop: 8,
    fontSize: 12,
    color: '#374151',
    fontWeight: '500',
  },
  section: {
    marginBottom: 20,
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 20,
    marginBottom: 12,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#111827',
  },
  sectionAction: {
    fontSize: 14,
    color: '#7c3aed',
    fontWeight: '500',
  },
  trackItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingVertical: 12,
    backgroundColor: 'white',
    marginHorizontal: 20,
    marginBottom: 8,
    borderRadius: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
    elevation: 1,
  },
  trackImage: {
    width: 50,
    height: 50,
    borderRadius: 8,
    marginRight: 12,
  },
  trackInfo: {
    flex: 1,
  },
  trackTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#111827',
    marginBottom: 4,
  },
  trackArtist: {
    fontSize: 14,
    color: '#6b7280',
  },
  trackActions: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  trackDuration: {
    fontSize: 12,
    color: '#6b7280',
    marginRight: 12,
  },
  playButton: {
    padding: 8,
  },
  playlistItem: {
    width: 150,
    marginRight: 16,
    backgroundColor: 'white',
    borderRadius: 12,
    padding: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
    elevation: 1,
  },
  playlistImage: {
    width: '100%',
    height: 120,
    borderRadius: 8,
    marginBottom: 8,
  },
  playlistInfo: {
    flex: 1,
  },
  playlistName: {
    fontSize: 14,
    fontWeight: '600',
    color: '#111827',
    marginBottom: 4,
  },
  playlistCount: {
    fontSize: 12,
    color: '#6b7280',
  },
});
