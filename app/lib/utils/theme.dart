import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

// ──────────────────────────────────────
// Color Palette — Minimal dark with claymorphic foundation
// ──────────────────────────────────────

class AppColors {
  AppColors._();

  // Primary accent — soft blue with subtle warmth
  static const primary = Color(0xFF6B8AFE);
  static const primaryLight = Color(0xFF8BA5FF);
  static const primaryDark = Color(0xFF4A6AE5);

  // Surface hierarchy — claymorphic depth
  static const background = Color(0xFF1A1B1E);
  static const surface = Color(0xFF242529);
  static const surfaceLight = Color(0xFF2E2F35);
  static const surfaceHighlight = Color(0xFF35363D);

  // Clay-specific surfaces for inner shadows
  static const clayBase = Color(0xFF242529);
  static const clayLight = Color(0xFF2E2F35);
  static const clayDark = Color(0xFF1C1C1F);

  // Semantic colors
  static const success = Color(0xFF4ADE80);
  static const error = Color(0xFFF87171);
  static const warning = Color(0xFFFBBF24);

  // Gradient pairs for accent effects (minimal use)
  static const gradientStart = Color(0xFF6B8AFE);
  static const gradientEnd = Color(0xFFA78BFA);

  // Text
  static const textPrimary = Color(0xFFF4F4F5);
  static const textSecondary = Color(0xFFA1A1AA);
  static const textMuted = Color(0xFF71717A);

  // Clay shadows
  static const clayShadowLight = Color(0xFF3A3B42);
  static const clayShadowDark = Color(0xFF141417);
}

// ──────────────────────────────────────
// Claymorphism Shadow Presets
// ──────────────────────────────────────

class AppClays {
  AppClays._();

  // Standard card clay - soft raised appearance
  static List<BoxShadow> card({double intensity = 1.0}) => [
    BoxShadow(
      color: AppColors.clayShadowLight.withValues(alpha: 0.15 * intensity),
      offset: const Offset(-4, -4),
      blurRadius: 12,
    ),
    BoxShadow(
      color: AppColors.clayShadowDark.withValues(alpha: 0.4 * intensity),
      offset: const Offset(6, 6),
      blurRadius: 16,
    ),
  ];

  // Pressed state - inverted shadows for inset effect
  static List<BoxShadow> pressed({double intensity = 1.0}) => [
    BoxShadow(
      color: AppColors.clayShadowDark.withValues(alpha: 0.3 * intensity),
      offset: const Offset(-2, -2),
      blurRadius: 6,
    ),
    BoxShadow(
      color: AppColors.clayShadowLight.withValues(alpha: 0.1 * intensity),
      offset: const Offset(3, 3),
      blurRadius: 8,
    ),
  ];

  // Button clay - medium elevation
  static List<BoxShadow> button({double intensity = 1.0}) => [
    BoxShadow(
      color: AppColors.clayShadowLight.withValues(alpha: 0.12 * intensity),
      offset: const Offset(-3, -3),
      blurRadius: 8,
    ),
    BoxShadow(
      color: AppColors.clayShadowDark.withValues(alpha: 0.35 * intensity),
      offset: const Offset(4, 4),
      blurRadius: 12,
    ),
  ];

  // Input field - subtle inset feel
  static List<BoxShadow> input({double intensity = 1.0}) => [
    BoxShadow(
      color: AppColors.clayShadowDark.withValues(alpha: 0.25 * intensity),
      offset: const Offset(2, 2),
      blurRadius: 6,
      spreadRadius: -1,
    ),
  ];

  // Icon container - floating appearance
  static List<BoxShadow> iconContainer({double intensity = 1.0}) => [
    BoxShadow(
      color: AppColors.clayShadowLight.withValues(alpha: 0.1 * intensity),
      offset: const Offset(-2, -2),
      blurRadius: 6,
    ),
    BoxShadow(
      color: AppColors.clayShadowDark.withValues(alpha: 0.3 * intensity),
      offset: const Offset(3, 3),
      blurRadius: 8,
    ),
  ];

  // Bottom nav - floating bar
  static List<BoxShadow> navBar({double intensity = 1.0}) => [
    BoxShadow(
      color: AppColors.clayShadowLight.withValues(alpha: 0.08 * intensity),
      offset: const Offset(-2, -2),
      blurRadius: 8,
    ),
    BoxShadow(
      color: AppColors.clayShadowDark.withValues(alpha: 0.35 * intensity),
      offset: const Offset(4, 4),
      blurRadius: 16,
    ),
  ];
}

// ──────────────────────────────────────
// Gradient Presets (minimal use - mostly solid colors)
// ──────────────────────────────────────

class AppGradients {
  AppGradients._();

  // Accent gradient - subtle, used sparingly
  static const accent = LinearGradient(
    colors: [AppColors.gradientStart, AppColors.gradientEnd],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );

  // Primary button fill for CTAs
  static const primaryButton = LinearGradient(
    colors: [AppColors.primary, AppColors.primaryDark],
    begin: Alignment.topCenter,
    end: Alignment.bottomCenter,
  );
}

// ──────────────────────────────────────
// App Theme
// ──────────────────────────────────────

class AppTheme {
  AppTheme._();

  static ThemeData get dark {
    final textTheme = GoogleFonts.interTextTheme(ThemeData.dark().textTheme);
    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.dark,
      primaryColor: AppColors.primary,
      scaffoldBackgroundColor: AppColors.background,
      cardColor: AppColors.surface,
      textTheme: textTheme,
      colorScheme: const ColorScheme.dark(
        primary: AppColors.primary,
        surface: AppColors.surface,
        error: AppColors.error,
        onSurface: AppColors.textPrimary,
      ),
      appBarTheme: AppBarTheme(
        backgroundColor: AppColors.background,
        surfaceTintColor: Colors.transparent,
        elevation: 0,
        titleTextStyle: GoogleFonts.inter(
          fontSize: 20,
          fontWeight: FontWeight.w700,
          color: AppColors.textPrimary,
          letterSpacing: -0.3,
        ),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: AppColors.surfaceLight,
          foregroundColor: AppColors.textPrimary,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
          elevation: 0,
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: AppColors.surface,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: BorderSide.none,
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: BorderSide.none,
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: const BorderSide(color: AppColors.primary, width: 1.5),
        ),
        labelStyle: const TextStyle(color: AppColors.textSecondary),
        hintStyle: const TextStyle(color: AppColors.textMuted),
      ),
      dividerTheme: const DividerThemeData(
        color: Color(0xFF2E2F35),
        thickness: 1,
      ),
      snackBarTheme: SnackBarThemeData(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        behavior: SnackBarBehavior.floating,
      ),
    );
  }
}

