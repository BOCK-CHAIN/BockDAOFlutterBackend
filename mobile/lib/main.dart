import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'providers/dao_provider.dart';
import 'screens/home_screen.dart';
import 'screens/wallet_screen.dart';
import 'screens/create_proposal_screen.dart';
import 'screens/treasury_screen.dart';

void main() {
  runApp(const BockDAOApp());
}

class BockDAOApp extends StatelessWidget {
  const BockDAOApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (context) => DAOProvider(),
      child: MaterialApp(
        title: 'Bock DAO',
        theme: ThemeData(
          colorScheme: ColorScheme.fromSeed(
            seedColor: const Color(0xFF1976D2), // ProjectX blue
            brightness: Brightness.light,
          ),
          useMaterial3: true,
          appBarTheme: const AppBarTheme(
            centerTitle: true,
            elevation: 2,
          ),
          cardTheme: CardThemeData(
            elevation: 2,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
            ),
          ),
          elevatedButtonTheme: ElevatedButtonThemeData(
            style: ElevatedButton.styleFrom(
              padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
          ),
          inputDecorationTheme: InputDecorationTheme(
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
            ),
            contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          ),
        ),
        darkTheme: ThemeData(
          colorScheme: ColorScheme.fromSeed(
            seedColor: const Color(0xFF1976D2),
            brightness: Brightness.dark,
          ),
          useMaterial3: true,
        ),
        home: const HomeScreen(),
        routes: {
          '/wallet': (context) => const WalletScreen(),
          '/create-proposal': (context) => const CreateProposalScreen(),
          '/treasury': (context) => const TreasuryScreen(),
        },
        debugShowCheckedModeBanner: false,
      ),
    );
  }
}