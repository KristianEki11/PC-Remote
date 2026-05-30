import 'package:flutter/material.dart';

/// Interactive scale animation wrapper for buttons.
/// Provides a subtle press-down effect for better tactile feedback.
class ScaleButtonWrapper extends StatefulWidget {
  final Widget child;

  const ScaleButtonWrapper({super.key, required this.child});

  @override
  State<ScaleButtonWrapper> createState() => _ScaleButtonWrapperState();
}

class _ScaleButtonWrapperState extends State<ScaleButtonWrapper> {
  bool _isPressed = false;

  @override
  Widget build(BuildContext context) {
    return Listener(
      onPointerDown: (_) => setState(() => _isPressed = true),
      onPointerUp: (_) => setState(() => _isPressed = false),
      onPointerCancel: (_) => setState(() => _isPressed = false),
      child: AnimatedScale(
        scale: _isPressed ? 0.95 : 1.0,
        duration: const Duration(milliseconds: 100),
        child: widget.child,
      ),
    );
  }
}
