import 'package:flutter/material.dart';
import '../utils/theme.dart';

class SharedCard extends StatelessWidget {
  final Widget child;
  final double borderRadius;
  final EdgeInsets? padding;

  const SharedCard({
    super.key,
    required this.child,
    this.borderRadius = 24,
    this.padding,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: AppGlass.cardDecoration().copyWith(
        borderRadius: BorderRadius.circular(borderRadius),
      ),
      child: Padding(
        padding: padding ?? const EdgeInsets.all(24.0),
        child: RepaintBoundary(child: child),
      ),
    );
  }
}

class CardHeader extends StatelessWidget {
  final IconData icon;
  final String title;
  final Widget? trailing;
  final Color? iconColor;

  const CardHeader({
    super.key,
    required this.icon,
    required this.title,
    this.trailing,
    this.iconColor,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.all(12),
          decoration: AppGlass.buttonDecoration().copyWith(
            borderRadius: BorderRadius.circular(14),
          ),
          child: Icon(
            icon,
            color: iconColor ?? AppColors.primary,
            size: 22,
          ),
        ),
        const SizedBox(width: 16),
        Text(
          title,
          style: const TextStyle(
            fontSize: 18,
            fontWeight: FontWeight.w700,
            color: AppColors.textPrimary,
            letterSpacing: -0.3,
          ),
        ),
        const Spacer(),
        if (trailing != null) trailing!,
      ],
    );
  }
}

class CardLoadingIndicator extends StatelessWidget {
  const CardLoadingIndicator({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(10),
      decoration: AppGlass.buttonDecoration().copyWith(
        borderRadius: BorderRadius.circular(12),
      ),
      child: const SizedBox(
        width: 18,
        height: 18,
        child: CircularProgressIndicator(
          strokeWidth: 2,
          color: AppColors.primary,
        ),
      ),
    );
  }
}
