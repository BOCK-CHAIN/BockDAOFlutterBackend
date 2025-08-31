# ProjectX DAO Mobile - Complete Deployment Guide

## 🎯 Overview

This guide provides complete instructions for deploying the ProjectX DAO mobile application, from building the APK to distributing it to end users.

## 📋 Prerequisites

### Development Environment
- Flutter SDK 3.27.x or higher
- Dart SDK 3.8.1 or higher
- Android SDK with API level 34
- Android NDK 27.0.12077973
- Java 11 or higher

### Server Requirements
- ProjectX DAO backend server running
- API accessible on port 9000
- WebSocket support enabled

## 🔨 Building the APK

### Quick Build (Automated)

#### Windows
```batch
cd projectx/mobile
build_apk.bat
```

#### Linux/Mac
```bash
cd projectx/mobile
chmod +x build_apk.sh
./build_apk.sh
```

### Manual Build Process

#### Step 1: Clean Previous Builds
```bash
flutter clean
```

#### Step 2: Install Dependencies
```bash
flutter pub get
```

#### Step 3: Generate Code
```bash
dart run build_runner build --delete-conflicting-outputs
```

#### Step 4: Build Debug APK
```bash
flutter build apk --debug
```

#### Step 5: Build Release APK
```bash
flutter build apk --release
```

### Build Outputs
- **Debug APK:** `build/app/outputs/flutter-apk/app-debug.apk`
- **Release APK:** `build/app/outputs/flutter-apk/app-release.apk`

## 📱 APK Distribution

### Internal Testing

#### Direct Installation
1. Transfer APK to Android device
2. Enable "Install from unknown sources"
3. Install APK file
4. Launch and test

#### ADB Installation
```bash
adb install build/app/outputs/flutter-apk/app-release.apk
```

### Production Distribution

#### Google Play Store (Future)
1. Create signed APK with release keystore
2. Upload to Google Play Console
3. Complete store listing
4. Submit for review

#### Direct Distribution
1. Host APK on secure server
2. Provide download link to users
3. Include installation instructions
4. Monitor usage and feedback

## 🔧 Configuration for Different Environments

### Development Environment
```dart
// lib/services/api_service.dart
static const String baseUrl = 'http://10.0.2.2:9000'; // Emulator
```

### Staging Environment
```dart
// lib/services/api_service.dart
static const String baseUrl = 'https://staging-api.projectx-dao.com';
```

### Production Environment
```dart
// lib/services/api_service.dart
static const String baseUrl = 'https://api.projectx-dao.com';
```

## 🔐 Security Configuration

### Release Signing

#### Generate Keystore
```bash
keytool -genkey -v -keystore projectx-dao-key.jks -keyalg RSA -keysize 2048 -validity 10000 -alias projectx-dao
```

#### Configure Gradle
```kotlin
// android/app/build.gradle.kts
android {
    signingConfigs {
        release {
            keyAlias = "projectx-dao"
            keyPassword = "your-key-password"
            storeFile = file("../projectx-dao-key.jks")
            storePassword = "your-store-password"
        }
    }
    buildTypes {
        release {
            signingConfig = signingConfigs.getByName("release")
        }
    }
}
```

### Network Security
```xml
<!-- android/app/src/main/res/xml/network_security_config.xml -->
<?xml version="1.0" encoding="utf-8"?>
<network-security-config>
    <domain-config cleartextTrafficPermitted="false">
        <domain includeSubdomains="true">api.projectx-dao.com</domain>
    </domain-config>
</network-security-config>
```

## 🚀 Deployment Checklist

### Pre-Deployment
- [ ] All features tested and working
- [ ] API endpoints configured correctly
- [ ] Security configurations in place
- [ ] Performance optimizations applied
- [ ] Error handling implemented
- [ ] User documentation prepared

### Build Process
- [ ] Clean build environment
- [ ] Dependencies updated
- [ ] Code generation completed
- [ ] Debug APK built and tested
- [ ] Release APK built and verified
- [ ] APK size optimized

### Post-Deployment
- [ ] Installation tested on multiple devices
- [ ] Network connectivity verified
- [ ] User feedback collected
- [ ] Performance monitoring enabled
- [ ] Update mechanism planned

## 📊 Monitoring and Analytics

### Crash Reporting
Consider integrating:
- Firebase Crashlytics
- Sentry
- Bugsnag

### Usage Analytics
Consider integrating:
- Firebase Analytics
- Google Analytics
- Custom analytics solution

### Performance Monitoring
- Monitor app startup time
- Track API response times
- Monitor memory usage
- Track user engagement

## 🔄 Update Strategy

### Over-the-Air Updates
- Implement update checking mechanism
- Provide in-app update notifications
- Support incremental updates

### Version Management
```yaml
# pubspec.yaml
version: 1.0.0+1  # version+build_number
```

### Release Notes
Maintain changelog for each version:
- New features
- Bug fixes
- Performance improvements
- Breaking changes

## 🧪 Testing Strategy

### Unit Testing
```bash
flutter test
```

### Integration Testing
```bash
flutter test integration_test/
```

### Device Testing
- Test on various Android versions
- Test different screen sizes
- Test network conditions
- Test edge cases

## 📋 Maintenance

### Regular Updates
- Security patches
- Dependency updates
- Performance improvements
- New features

### Monitoring
- Crash rates
- User feedback
- Performance metrics
- Usage statistics

### Support
- User documentation
- FAQ section
- Support channels
- Bug reporting

## 🔧 Troubleshooting

### Build Issues

#### Gradle Build Fails
```bash
cd android
./gradlew clean
cd ..
flutter clean
flutter pub get
```

#### NDK Version Mismatch
Update `android/app/build.gradle.kts`:
```kotlin
android {
    ndkVersion = "27.0.12077973"
}
```

#### Dependency Conflicts
```bash
flutter pub deps
flutter pub upgrade
```

### Runtime Issues

#### Network Connection
- Check API server status
- Verify network permissions
- Test with different networks

#### Wallet Connection
- Validate private key format
- Check address format
- Verify server connectivity

#### Performance Issues
- Enable performance profiling
- Check memory usage
- Optimize image loading
- Reduce bundle size

## 📞 Support and Resources

### Documentation
- Flutter documentation: https://flutter.dev/docs
- Android development: https://developer.android.com
- ProjectX DAO documentation: See main README.md

### Community
- GitHub Issues: Report bugs and feature requests
- Discussions: Community support and questions
- Discord/Telegram: Real-time community chat

### Professional Support
- Code reviews
- Security audits
- Performance optimization
- Custom feature development

---

**Deployment Guide Version:** 1.0  
**Last Updated:** August 28, 2025  
**Compatible with:** Flutter 3.27.x, Android API 21-34