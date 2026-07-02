(function () {
  function ready(fn) {
    if (document.readyState === "loading") {
      document.addEventListener("DOMContentLoaded", fn, { once: true });
    } else {
      fn();
    }
  }

  function mat(scene, name, color, options = {}) {
    const material = new BABYLON.StandardMaterial(name, scene);
    material.diffuseColor = color;
    material.emissiveColor = options.glow ? color.scale(options.glow) : color.scale(0.06);
    material.specularColor = options.specular || new BABYLON.Color3(0.78, 0.68, 0.42);
    material.alpha = options.alpha ?? 1;
    material.backFaceCulling = options.backFaceCulling ?? true;
    return material;
  }

  function makeDiscTexture(scene) {
    const texture = new BABYLON.DynamicTexture("avatarParticleDot", { width: 64, height: 64 }, scene);
    const ctx = texture.getContext();
    const gradient = ctx.createRadialGradient(32, 32, 0, 32, 32, 31);
    gradient.addColorStop(0, "rgba(255,246,206,1)");
    gradient.addColorStop(0.42, "rgba(91,221,255,0.68)");
    gradient.addColorStop(1, "rgba(91,221,255,0)");
    ctx.fillStyle = gradient;
    ctx.fillRect(0, 0, 64, 64);
    texture.update();
    return texture;
  }

  function createAvatar(scene, materials) {
    const root = new BABYLON.TransformNode("playerAvatarRoot", scene);

    const coat = BABYLON.MeshBuilder.CreateCylinder("avatarCoat", {
      diameterTop: 0.72,
      diameterBottom: 1.22,
      height: 1.72,
      tessellation: 8,
    }, scene);
    coat.position.y = 0.03;
    coat.material = materials.coat;
    coat.parent = root;

    const chest = BABYLON.MeshBuilder.CreateBox("avatarChestArmor", {
      width: 1.06,
      height: 0.82,
      depth: 0.28,
    }, scene);
    chest.position.set(0, 0.36, -0.11);
    chest.rotation.x = -0.08;
    chest.material = materials.armor;
    chest.parent = root;

    const core = BABYLON.MeshBuilder.CreateSphere("avatarCoreBadge", { diameter: 0.24, segments: 24 }, scene);
    core.position.set(0, 0.42, -0.28);
    core.material = materials.cyan;
    core.parent = root;

    const head = BABYLON.MeshBuilder.CreateSphere("avatarHead", {
      diameterX: 0.68,
      diameterY: 0.78,
      diameterZ: 0.64,
      segments: 40,
    }, scene);
    head.position.y = 1.22;
    head.material = materials.skin;
    head.parent = root;

    const hair = BABYLON.MeshBuilder.CreateSphere("avatarHair", {
      diameterX: 0.76,
      diameterY: 0.52,
      diameterZ: 0.7,
      segments: 32,
    }, scene);
    hair.position.set(0, 1.48, -0.03);
    hair.scaling.y = 0.72;
    hair.material = materials.hair;
    hair.parent = root;

    const visor = BABYLON.MeshBuilder.CreateBox("avatarVisor", {
      width: 0.56,
      height: 0.09,
      depth: 0.035,
    }, scene);
    visor.position.set(0, 1.28, -0.33);
    visor.material = materials.cyan;
    visor.parent = root;

    [-1, 1].forEach((side) => {
      const shoulder = BABYLON.MeshBuilder.CreateSphere(`avatarShoulder${side}`, {
        diameterX: 0.44,
        diameterY: 0.26,
        diameterZ: 0.38,
        segments: 24,
      }, scene);
      shoulder.position.set(side * 0.7, 0.54, -0.03);
      shoulder.rotation.z = side * 0.3;
      shoulder.material = materials.gold;
      shoulder.parent = root;

      const arm = BABYLON.MeshBuilder.CreateCylinder(`avatarArm${side}`, {
        diameterTop: 0.2,
        diameterBottom: 0.24,
        height: 0.9,
        tessellation: 12,
      }, scene);
      arm.position.set(side * 0.86, 0.02, -0.02);
      arm.rotation.z = side * 0.22;
      arm.material = materials.coat;
      arm.parent = root;

      const blade = BABYLON.MeshBuilder.CreateBox(`avatarEnergyBlade${side}`, {
        width: 0.035,
        height: 1.15,
        depth: 0.06,
      }, scene);
      blade.position.set(side * 1.15, 0.36, 0.18);
      blade.rotation.z = side * -0.18;
      blade.material = materials.cyan;
      blade.parent = root;
    });

    const scarf = BABYLON.MeshBuilder.CreateTorus("avatarScarfRing", {
      diameter: 0.96,
      thickness: 0.045,
      tessellation: 80,
    }, scene);
    scarf.position.y = 0.83;
    scarf.rotation.x = Math.PI * 0.5;
    scarf.material = materials.red;
    scarf.parent = root;

    const cloak = BABYLON.MeshBuilder.CreateCylinder("avatarCloak", {
      diameterTop: 0.64,
      diameterBottom: 1.46,
      height: 1.5,
      tessellation: 5,
    }, scene);
    cloak.position.set(0, -0.08, 0.21);
    cloak.rotation.x = 0.12;
    cloak.material = materials.cloak;
    cloak.parent = root;

    return root;
  }

  function createScene(engine, canvas) {
    const scene = new BABYLON.Scene(engine);
    scene.clearColor = new BABYLON.Color4(0.01, 0.05, 0.09, 0);

    const camera = new BABYLON.ArcRotateCamera(
      "avatarCamera",
      Math.PI * 1.12,
      Math.PI * 0.46,
      7.8,
      new BABYLON.Vector3(0, 0.42, 0),
      scene,
    );
    camera.attachControl(canvas, true);
    camera.lowerRadiusLimit = 6.4;
    camera.upperRadiusLimit = 9.2;
    camera.inputs.attached.mousewheel.detachControl();

    const hemi = new BABYLON.HemisphericLight("avatarSky", new BABYLON.Vector3(0, 1, 0), scene);
    hemi.intensity = 0.52;
    hemi.groundColor = new BABYLON.Color3(0.02, 0.08, 0.14);

    const key = new BABYLON.DirectionalLight("avatarKey", new BABYLON.Vector3(-0.45, -0.72, 0.4), scene);
    key.position = new BABYLON.Vector3(4, 7, -5);
    key.intensity = 1.85;

    const rim = new BABYLON.PointLight("avatarRim", new BABYLON.Vector3(-2.6, 1.7, -2.2), scene);
    rim.diffuse = new BABYLON.Color3(0.36, 0.82, 1);
    rim.intensity = 1.3;

    const glow = new BABYLON.GlowLayer("avatarGlow", scene);
    glow.intensity = 0.72;

    const materials = {
      armor: mat(scene, "avatarArmor", new BABYLON.Color3(0.18, 0.28, 0.36), { specular: new BABYLON.Color3(0.9, 0.78, 0.44) }),
      cloak: mat(scene, "avatarCloakMat", new BABYLON.Color3(0.04, 0.11, 0.19), { alpha: 0.86, backFaceCulling: false }),
      coat: mat(scene, "avatarCoatMat", new BABYLON.Color3(0.08, 0.18, 0.29)),
      cyan: mat(scene, "avatarCyanGlow", new BABYLON.Color3(0.27, 0.88, 1), { glow: 0.82 }),
      gold: mat(scene, "avatarGold", new BABYLON.Color3(1, 0.72, 0.27), { glow: 0.24 }),
      hair: mat(scene, "avatarHairMat", new BABYLON.Color3(0.05, 0.07, 0.1)),
      red: mat(scene, "avatarAccentRed", new BABYLON.Color3(0.9, 0.18, 0.24), { glow: 0.2 }),
      skin: mat(scene, "avatarSkin", new BABYLON.Color3(0.82, 0.64, 0.48)),
    };

    const avatar = createAvatar(scene, materials);

    const platform = BABYLON.MeshBuilder.CreateCylinder("avatarPlatform", {
      diameterTop: 2.82,
      diameterBottom: 3.62,
      height: 0.18,
      tessellation: 10,
    }, scene);
    platform.position.y = -1.06;
    platform.material = mat(scene, "avatarPlatformMat", new BABYLON.Color3(0.05, 0.14, 0.23), { alpha: 0.72, backFaceCulling: false });

    const ringA = BABYLON.MeshBuilder.CreateTorus("avatarDataRingA", {
      diameter: 3.92,
      thickness: 0.026,
      tessellation: 128,
    }, scene);
    ringA.position.y = -0.9;
    ringA.rotation.x = Math.PI * 0.5;
    ringA.material = materials.gold;

    const ringB = BABYLON.MeshBuilder.CreateTorus("avatarDataRingB", {
      diameter: 2.28,
      thickness: 0.02,
      tessellation: 96,
    }, scene);
    ringB.position.y = 0.52;
    ringB.rotation.x = Math.PI * 0.5;
    ringB.rotation.y = 0.2;
    ringB.material = materials.cyan;

    const scanLines = [];
    for (let i = 0; i < 6; i += 1) {
      const line = BABYLON.MeshBuilder.CreateTorus(`avatarScanLine${i}`, {
        diameter: 1.44 + i * 0.18,
        thickness: 0.01,
        tessellation: 96,
      }, scene);
      line.position.y = -0.54 + i * 0.28;
      line.rotation.x = Math.PI * 0.5;
      line.material = i % 2 ? materials.cyan : materials.gold;
      scanLines.push(line);
    }

    const markerMaterial = mat(scene, "avatarMarkerGlow", new BABYLON.Color3(0.38, 0.9, 1), { glow: 0.7 });
    const markers = [];
    for (let i = 0; i < 12; i += 1) {
      const angle = (Math.PI * 2 * i) / 12;
      const marker = BABYLON.MeshBuilder.CreateBox(`avatarMarker${i}`, {
        width: 0.06,
        height: 0.18,
        depth: 0.06,
      }, scene);
      marker.position.set(Math.cos(angle) * 2.18, -0.86, Math.sin(angle) * 2.18);
      marker.rotation.y = -angle;
      marker.material = markerMaterial;
      markers.push(marker);
    }

    const particles = new BABYLON.ParticleSystem("avatarLightDust", 900, scene);
    particles.particleTexture = makeDiscTexture(scene);
    particles.emitter = new BABYLON.Vector3(0, 0.15, 0);
    particles.minEmitBox = new BABYLON.Vector3(-2.2, -0.72, -1.6);
    particles.maxEmitBox = new BABYLON.Vector3(2.2, 1.82, 1.6);
    particles.color1 = new BABYLON.Color4(1, 0.74, 0.26, 0.68);
    particles.color2 = new BABYLON.Color4(0.28, 0.82, 1, 0.6);
    particles.colorDead = new BABYLON.Color4(0.02, 0.06, 0.1, 0);
    particles.minSize = 0.018;
    particles.maxSize = 0.09;
    particles.minLifeTime = 1.4;
    particles.maxLifeTime = 3.8;
    particles.emitRate = 88;
    particles.blendMode = BABYLON.ParticleSystem.BLENDMODE_ADD;
    particles.gravity = new BABYLON.Vector3(0, 0.018, 0);
    particles.direction1 = new BABYLON.Vector3(-0.06, 0.14, -0.06);
    particles.direction2 = new BABYLON.Vector3(0.06, 0.28, 0.06);
    particles.start();

    scene.onBeforeRenderObservable.add(() => {
      const dt = engine.getDeltaTime() * 0.001;
      const t = performance.now() * 0.001;
      avatar.rotation.y += dt * 0.18;
      avatar.position.y = Math.sin(t * 1.35) * 0.035;
      ringA.rotation.z -= dt * 0.34;
      ringB.rotation.z += dt * 0.58;
      platform.rotation.y += dt * 0.14;
      camera.alpha += dt * 0.018;
      scanLines.forEach((line, index) => {
        line.scaling.x = 1 + Math.sin(t * 1.7 + index) * 0.018;
        line.scaling.y = 1 + Math.cos(t * 1.5 + index) * 0.018;
      });
      markers.forEach((marker, index) => {
        marker.position.y = -0.84 + Math.sin(t * 1.8 + index) * 0.045;
      });
    });

    return scene;
  }

  ready(() => {
    const canvas = document.getElementById("mypageBabylonCanvas");
    if (!canvas || !window.BABYLON) {
      document.body.classList.add("mypage-3d-fallback");
      return;
    }

    const engine = new BABYLON.Engine(canvas, true, {
      antialias: true,
      adaptToDeviceRatio: true,
      preserveDrawingBuffer: true,
      stencil: true,
    });
    const scene = createScene(engine, canvas);
    engine.runRenderLoop(() => scene.render());
    window.addEventListener("resize", () => engine.resize());
    document.body.classList.add("mypage-3d-ready", "mypage-avatar-ready");
  });
})();
