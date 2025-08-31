# ProjectX DAO Mobile APK - Build Information

## 📱 Application Details

**App Name:** Bock DAO  
**Package Name:** com.projectx.dao.bock_dao_mobile  
**Version:** 1.0.0+1  
**Build Date:** August 28, 2025  
**Flutter Version:** 3.27.x  
**Target SDK:** Android API 34  
**Minimum SDK:** Android API 21 (Android 5.0)  

## 📦 APK Files

### Debug APK
- **File:** `build/app/outputs/flutter-apk/app-debug.apk`
- **Size:** ~25MB
- **Purpose:** Development and testing
- **Features:** Debug symbols, logging enabled
- **Installation:** Can be installed on any Android device with "Unknown sources" enabled

### Release APK
- **File:** `build/app/outputs/flutter-apk/app-release.apk`
- **Size:** 21.7MB
- **Purpose:** Production deployment
- **Features:** Optimized, tree-shaken, minified
- **Installation:** Ready for distribution

## 🚀 Features Included

### Core DAO Features
- ✅ **Wallet Integration**: Manual wallet connection with private key support
- ✅ **Proposal Management**: View, create, and manage governance proposals
- ✅ **Voting System**: Cast votes on active proposals with different voting types
- ✅ **Treasury Dashboard**: View treasury status and fund management
- ✅ **Real-time Updates**: WebSocket integration for live updates
- ✅ **Secure Storage**: Encrypted storage for wallet credentials

### User Interface
- ✅ **Material Design 3**: Modern, responsive UI design
- ✅ **Dark/Light Theme**: Automatic theme switching support
- ✅ **Navigation**: Bottom navigation with 4 main sections
- ✅ **Responsive Layout**: Optimized for different screen sizes
- ✅ **Loading States**: Proper loading indicators and error handling

### Technical Features
- ✅ **HTTP API Integration**: RESTful API communication
- ✅ **WebSocket Support**: Real-time event streaming
- ✅ **JSON Serialization**: Efficient data parsing
- ✅ **Secure Storage**: Flutter secure storage for sensitive data
- ✅ **State Management**: Provider pattern for app state
- ✅ **Error Handling**: Comprehensive error management

## 🔧 Installation Instructions

### Prerequisites
- Android device running Android 5.0 (API 21) or higher
- At least 50MB of free storage space
- Internet connection for API communication

### Installation Steps

#### Method 1: Direct APK Installation
1. Download the APK file to your Android device
2. Enable "Install from unknown sources" in Settings > Security
3. Open the APK file and tap "Install"
4. Grant necessary permissions when prompted
5. Launch "Bock DAO" from your app drawer

#### Method 2: ADB Installation (Developer)
```bash
# Connect your device via USB with USB debugging enabled
adb install build/app/outputs/flutter-apk/app-release.apk
```

## 🌐 Network Configuration

### API Endpoints
- **Base URL:** `http://10.0.2.2:9000` (Android Emulator)
- **WebSocket:** `ws://10.0.2.2:9000/dao/events`

### For Physical Devices
If installing on a physical device, you may need to update the API endpoints in the app to point to your actual server IP address instead of the emulator localhost (10.0.2.2).

## 📋 Permissions Required

### Internet Permissions
- `android.permission.INTERNET` - Required for API communication
- `android.permission.ACCESS_NETWORK_STATE` - Required for network status

### Storage Permissions
- Secure storage access for wallet credentials (handled by Flutter Secure Storage)

## 🔐 Security Features

### Wallet Security
- Private keys stored in Android Keystore via Flutter Secure Storage
- Keys encrypted at rest
- No private key transmission over network
- Session-based authentication

### Network Security
- HTTPS support (when server supports it)
- Certificate pinning ready (can be configured)
- Request/response validation
- Error message sanitization

## 🧪 Testing

### Tested Scenarios
- ✅ App installation and launch
- ✅ Wallet connection with manual private key
- ✅ Proposal viewing and filtering
- ✅ Navigation between screens
- ✅ Error handling for network issues
- ✅ State persistence across app restarts

### Test Devices
- Android Emulator (API 34)
- Various screen sizes and orientations

## 🚀 Usage Guide

### First Launch
1. Open the Bock DAO app
2. Navigate to the Wallet tab (bottom navigation)
3. Tap "Connect Wallet"
4. Choose "Manual Wallet" option
5. Enter your private key and address
6. Tap "Connect"

### Creating Proposals
1. Navigate to the "Create" tab
2. Fill in proposal details:
   - Title
   - Description
   - Type (General, Treasury, Technical, Parameter)
   - Voting type
   - Duration
3. Tap "Create Proposal"

### Voting on Proposals
1. Navigate to the "Proposals" tab
2. Tap on any active proposal
3. Review proposal details
4. Select your vote choice (Yes/No/Abstain)
5. Tap "Cast Vote"

### Treasury Management
1. Navigate to the "Treasury" tab
2. View current treasury balance
3. See recent transactions
4. Monitor fund allocation

## 🔧 Troubleshooting

### Common Issues

#### App Won't Install
- Ensure "Unknown sources" is enabled
- Check available storage space
- Verify Android version compatibility

#### Network Connection Issues
- Check internet connection
- Verify server is running on correct port
- For physical devices, update API endpoints

#### Wallet Connection Fails
- Verify private key format (64 hex characters)
- Check address format (40 hex characters with 0x prefix)
- Ensure server is accessible

### Debug Information
- Enable debug logging in developer options
- Check logcat for detailed error messages
- Verify API server status

## 📞 Support

### Getting Help
- Check the troubleshooting section above
- Review the main README.md for server setup
- Open GitHub issues for bugs
- Join community discussions

### Development
- Source code: `projectx/mobile/`
- Build scripts: `build_apk.bat` (Windows) or `build_apk.sh` (Linux/Mac)
- Flutter version: Check `pubspec.yaml`

## 📄 License

This mobile application is part of the Bock DAO project and follows the same MIT license terms as the main project.

---

**Built with ❤️ using Flutter**  
*APK generated on: August 28, 2025*  
*Flutter SDK: 3.27.x*  
*Dart SDK: 3.8.1*