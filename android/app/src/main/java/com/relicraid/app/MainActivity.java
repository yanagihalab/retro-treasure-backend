package com.relicraid.app;

import android.annotation.SuppressLint;
import android.app.Activity;
import android.graphics.Color;
import android.net.http.SslError;
import android.os.Bundle;
import android.view.Gravity;
import android.view.View;
import android.view.Window;
import android.view.WindowInsets;
import android.view.WindowInsetsController;
import android.webkit.SslErrorHandler;
import android.webkit.WebChromeClient;
import android.webkit.WebResourceError;
import android.webkit.WebResourceRequest;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.widget.FrameLayout;
import android.widget.TextView;

public class MainActivity extends Activity {
    private WebView webView;
    private TextView errorView;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        requestWindowFeature(Window.FEATURE_NO_TITLE);

        FrameLayout root = new FrameLayout(this);
        root.setBackgroundColor(Color.rgb(3, 13, 24));

        webView = new WebView(this);
        errorView = buildErrorView();

        root.addView(webView, new FrameLayout.LayoutParams(
            FrameLayout.LayoutParams.MATCH_PARENT,
            FrameLayout.LayoutParams.MATCH_PARENT
        ));
        root.addView(errorView, new FrameLayout.LayoutParams(
            FrameLayout.LayoutParams.MATCH_PARENT,
            FrameLayout.LayoutParams.MATCH_PARENT
        ));

        setContentView(root);
        hideSystemUi();
        configureWebView();
        webView.loadUrl(BuildConfig.GAME_BASE_URL);
    }

    @Override
    public void onWindowFocusChanged(boolean hasFocus) {
        super.onWindowFocusChanged(hasFocus);
        if (hasFocus) hideSystemUi();
    }

    @Override
    public void onBackPressed() {
        if (webView != null && webView.canGoBack()) {
            webView.goBack();
            return;
        }
        super.onBackPressed();
    }

    @SuppressLint("SetJavaScriptEnabled")
    private void configureWebView() {
        WebSettings settings = webView.getSettings();
        settings.setJavaScriptEnabled(true);
        settings.setDomStorageEnabled(true);
        settings.setDatabaseEnabled(true);
        settings.setLoadWithOverviewMode(true);
        settings.setUseWideViewPort(true);
        settings.setMediaPlaybackRequiresUserGesture(false);
        settings.setMixedContentMode(WebSettings.MIXED_CONTENT_ALWAYS_ALLOW);
        String userAgent = settings.getUserAgentString();
        if (userAgent == null) {
            userAgent = "";
        }
        if (!userAgent.contains("RelicRaidAndroid")) {
            settings.setUserAgentString(userAgent + " RelicRaidAndroid/0.1");
        }

        webView.setBackgroundColor(Color.TRANSPARENT);
        webView.setWebChromeClient(new WebChromeClient());
        webView.setWebViewClient(new WebViewClient() {
            @Override
            public boolean shouldOverrideUrlLoading(WebView view, WebResourceRequest request) {
                return false;
            }

            @Override
            public void onPageFinished(WebView view, String url) {
                errorView.setVisibility(View.GONE);
            }

            @Override
            public void onReceivedError(WebView view, WebResourceRequest request, WebResourceError error) {
                if (request.isForMainFrame()) {
                    showConnectionError();
                }
            }

            @Override
            public void onReceivedSslError(WebView view, SslErrorHandler handler, SslError error) {
                handler.cancel();
                showConnectionError();
            }
        });
    }

    private TextView buildErrorView() {
        TextView view = new TextView(this);
        view.setVisibility(View.GONE);
        view.setGravity(Gravity.CENTER);
        view.setPadding(36, 24, 36, 24);
        view.setTextColor(Color.rgb(248, 252, 255));
        view.setTextSize(16);
        view.setBackgroundColor(Color.rgb(3, 13, 24));
        return view;
    }

    private void showConnectionError() {
        errorView.setText(
            "RELIC RAID に接続できません。\n\n" +
            "開発中はバックエンドを起動してから再度開いてください。\n" +
            "Emulator: http://10.0.2.2:8080/static/\n" +
            "Current: " + BuildConfig.GAME_BASE_URL
        );
        errorView.setVisibility(View.VISIBLE);
    }

    private void hideSystemUi() {
        Window window = getWindow();
        if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.R) {
            window.setDecorFitsSystemWindows(false);
            WindowInsetsController controller = window.getDecorView().getWindowInsetsController();
            if (controller != null) {
                controller.hide(WindowInsets.Type.statusBars() | WindowInsets.Type.navigationBars());
                controller.setSystemBarsBehavior(
                    WindowInsetsController.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
                );
            }
        } else {
            window.getDecorView().setSystemUiVisibility(
                View.SYSTEM_UI_FLAG_FULLSCREEN
                    | View.SYSTEM_UI_FLAG_HIDE_NAVIGATION
                    | View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY
                    | View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN
                    | View.SYSTEM_UI_FLAG_LAYOUT_HIDE_NAVIGATION
                    | View.SYSTEM_UI_FLAG_LAYOUT_STABLE
            );
        }
    }
}
