import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class AppColors {
  AppColors._();

  static const primary = Color(0xFF171717); // Graphite
  static const primaryLight = Color(0xFF404040);
  static const primaryDark = Color(0xFF000000);

  static const background = Color(0xFFFAFAFA); // Off-White

  static const textPrimary = Color(0xFF0A0A0A);
  static const textSecondary = Color(0xFF525252);
  static const textMuted = Color(0xFFA3A3A3);

  // Fallbacks for compatibility
  static const surface = Colors.white;
  static const surfaceLight = Color(0xFFF5F5F5);
  static const success = Color(0xFF3B82F6); // Apple Blue
  static const error = Color(0xFFEF4444);
  static const gradientEnd = Color(0xFF737373);
}

class AppGradients {
  AppGradients._();

  static const primaryButton = LinearGradient(
    colors: [AppColors.primaryLight, AppColors.primary],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );

  static const glassBackground = LinearGradient(
    colors: [Color(0xFFFAFAFA), Color(0xFFF5F5F5)],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );
}

class AppGlass {
  static BoxDecoration cardDecoration() {
    return BoxDecoration(
      color: Colors.white.withOpacity(0.95),
      borderRadius: BorderRadius.circular(24),
      border: Border.all(color: const Color(0xFFE5E5E5), width: 0.5),
      boxShadow: [
        BoxShadow(
          color: Colors.black.withOpacity(0.04),
          blurRadius: 24,
          offset: const Offset(0, 8),
        ),
      ],
    );
  }

  static BoxDecoration buttonDecoration() {
    return BoxDecoration(
      color: Colors.white.withOpacity(0.8),
      borderRadius: BorderRadius.circular(16),
      border: Border.all(color: const Color(0xFFE5E5E5), width: 0.5),
      boxShadow: [
        BoxShadow(
          color: Colors.black.withOpacity(0.03),
          blurRadius: 10,
          offset: const Offset(0, 4),
        ),
      ],
    );
  }
}

class AppClays {
  AppClays._();
  
  static List<BoxShadow> card({double intensity = 1.0}) => [];
  static List<BoxShadow> iconContainer({double intensity = 1.0}) => [];
  static List<BoxShadow> button({double intensity = 1.0}) => [];
  static List<BoxShadow> navBar({double intensity = 1.0}) => [];
}

class AppTheme {
  AppTheme._();

  // Redirect to light theme
  static ThemeData get dark => light;

  static ThemeData get light {
    final textTheme = GoogleFonts.interTextTheme(ThemeData.light().textTheme);
    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.light,
      primaryColor: AppColors.primary,
      scaffoldBackgroundColor: AppColors.background,
      textTheme: textTheme,
      colorScheme: const ColorScheme.light(
        primary: AppColors.primary,
        surface: Colors.white,
        error: AppColors.error,
        onSurface: AppColors.textPrimary,
      ),
      appBarTheme: AppBarTheme(
        backgroundColor: Colors.transparent,
        elevation: 0,
        centerTitle: true,
        titleTextStyle: GoogleFonts.inter(
          fontSize: 20,
          fontWeight: FontWeight.w700,
          color: AppColors.textPrimary,
          letterSpacing: -0.5,
        ),
        iconTheme: const IconThemeData(color: AppColors.primary),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: AppColors.primary,
          foregroundColor: Colors.white,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
          elevation: 0,
          minimumSize: const Size(88, 56), // Larger touch targets
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: Colors.white.withOpacity(0.8),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: BorderSide(color: const Color(0xFFE5E5E5), width: 0.5),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: BorderSide(color: const Color(0xFFE5E5E5), width: 0.5),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: const BorderSide(color: AppColors.primary, width: 1.5),
        ),
        labelStyle: const TextStyle(color: AppColors.textMuted),
        hintStyle: const TextStyle(color: AppColors.textMuted),
      ),
    );
  }
}
