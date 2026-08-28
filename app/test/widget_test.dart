import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:provider/provider.dart';
import 'package:pc_remote/main.dart';
import 'package:pc_remote/models/app_state.dart';
import 'package:pc_remote/models/audio_state.dart';
import 'package:pc_remote/models/media_state.dart';
import 'package:pc_remote/services/api_service.dart';

void main() {
  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await ApiService.init();
  });

  testWidgets('Login screen loads and displays PC Remote title', (WidgetTester tester) async {
    await tester.pumpWidget(
      MultiProvider(
        providers: [
          ChangeNotifierProvider(create: (_) => AppState()),
          ChangeNotifierProvider(create: (_) => AudioState()),
          ChangeNotifierProvider(create: (_) => MediaState()),
        ],
        child: const MyApp(),
      ),
    );
    await tester.pumpAndSettle();

    // Verify PC Remote brand text exists
    expect(find.text('PC Remote'), findsOneWidget);
    expect(find.text('Kontrol PC dari genggaman tangan'), findsOneWidget);

    // Verify QR scan button exists
    expect(find.text('Pindai QR Code di PC'), findsOneWidget);

    // Verify manual button exists
    expect(find.text('Hubungkan Manual'), findsOneWidget);
  });
}
